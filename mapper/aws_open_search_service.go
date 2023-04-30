package mapper

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/elasticsearchservice"
	goCache "github.com/patrickmn/go-cache"

	"github.com/terakoya76/modd/datadog"
)

const awsOpenSearchServiceCacheKey string = string(datadog.AwsOpenSearchService)

// AwsOpenSearchServiceClient is abstract interface of *elasticsearchservice.Client.
type AwsOpenSearchServiceClient interface {
	ListDomainNames(
		ctx context.Context,
		params *elasticsearchservice.ListDomainNamesInput,
		optFns ...func(*elasticsearchservice.Options),
	) (*elasticsearchservice.ListDomainNamesOutput, error)
	DescribeElasticsearchDomains(
		ctx context.Context,
		params *elasticsearchservice.DescribeElasticsearchDomainsInput,
		optFns ...func(*elasticsearchservice.Options),
	) (*elasticsearchservice.DescribeElasticsearchDomainsOutput, error)
	ListTags(
		ctx context.Context,
		params *elasticsearchservice.ListTagsInput,
		optFns ...func(*elasticsearchservice.Options),
	) (*elasticsearchservice.ListTagsOutput, error)
}

// AwsOpenSearchServiceTagsMapper implements TagsMapper for AWS OpenSearch Service.
type AwsOpenSearchServiceTagsMapper struct {
	cache  *goCache.Cache
	client AwsOpenSearchServiceClient
}

// BuildAwsOpenSearchServiceTagsMapper builds AwsOpenSearchServiceTagsMapper from args.
func BuildAwsOpenSearchServiceTagsMapper(cache *goCache.Cache, client AwsOpenSearchServiceClient) AwsOpenSearchServiceTagsMapper {
	return AwsOpenSearchServiceTagsMapper{
		cache:  cache,
		client: client,
	}
}

// GetAwsOpenSearchServiceClient returns AWS OpenSearch Service client.
func GetAwsOpenSearchServiceClient(ctx context.Context) (*elasticsearchservice.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRetryer(func() aws.Retryer {
		return retry.AddWithMaxAttempts(retry.NewStandard(), 10)
	}))
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return elasticsearchservice.NewFromConfig(cfg), nil
}

// GetTagsMapping returns the latest tags mapping.
func (tm AwsOpenSearchServiceTagsMapper) GetTagsMapping(ctx context.Context) (map[string]Tags, error) {
	if cv, found := tm.cache.Get(awsOpenSearchServiceCacheKey); found {
		mapping := cv.(map[string]Tags)
		return mapping, nil
	}

	mapping := make(map[string]Tags)

	domainsInput := elasticsearchservice.ListDomainNamesInput{}
	domainsOutput, err := tm.client.ListDomainNames(ctx, &domainsInput)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/elasticsearchservice#DescribeElasticsearchDomainsInput
	maxItemsPerReq := 5

	// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/elasticsearchservice#DescribeElasticsearchDomainsInput
	for i := 0; i < len(domainsOutput.DomainNames); i += 5 {
		domainNamesCap := len(domainsOutput.DomainNames) % maxItemsPerReq
		domainNames := make([]string, 0)
		for j := 0; j < domainNamesCap; j++ {
			domain := domainsOutput.DomainNames[i+j].DomainName
			domainNames = append(domainNames, *domain)
		}

		input := elasticsearchservice.DescribeElasticsearchDomainsInput{
			DomainNames: domainNames,
		}
		output, err := tm.client.DescribeElasticsearchDomains(ctx, &input)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		for j := 0; j < len(output.DomainStatusList); j++ {
			domain := output.DomainStatusList[j]
			tagsInput := elasticsearchservice.ListTagsInput{ARN: domain.ARN}
			tagsOutput, err := tm.client.ListTags(ctx, &tagsInput)
			if err != nil {
				return nil, fmt.Errorf("%w", err)
			}

			tags := make(Tags, len(tagsOutput.TagList))
			for k, tag := range tagsOutput.TagList {
				tags[k] = fmt.Sprintf("%s:%s", strings.ToLower(*tag.Key), strings.ToLower(*tag.Value))
			}
			mapping[*domain.DomainName] = tags
		}
	}

	tm.cache.Set(awsOpenSearchServiceCacheKey, mapping, goCache.DefaultExpiration)
	return mapping, nil
}
