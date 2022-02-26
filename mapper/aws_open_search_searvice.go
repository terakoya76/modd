package mapper

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/elasticsearchservice"
	"github.com/patrickmn/go-cache"

	"github.com/terakoya76/modd/datadog"
)

const awsOpenSearchServiceCacheKey string = string(datadog.AwsOpenSearchService)

// AwsOpenSearchServiceTagsMapper implements TagsMapper for AWS OpenSearch Service.
type AwsOpenSearchServiceTagsMapper struct {
	cache  *cache.Cache
	client *elasticsearchservice.Client
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

	maxItemsPerReq := 5

	// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/elasticsearchservice#DescribeElasticsearchDomainsInput
	domainNames := make([]string, maxItemsPerReq)
	for i := 0; i < len(domainsOutput.DomainNames); i++ {
		idx := (i + 1) % maxItemsPerReq
		domainNames[idx] = *domainsOutput.DomainNames[i].DomainName

		if idx != 0 {
			continue
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

	tm.cache.Set(awsOpenSearchServiceCacheKey, mapping, cache.DefaultExpiration)
	return mapping, nil
}
