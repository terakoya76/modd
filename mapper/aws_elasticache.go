package mapper

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	goCache "github.com/patrickmn/go-cache"

	"github.com/terakoya76/modd/datadog"
)

const awsElastiCacheCacheKey string = string(datadog.AwsElastiCache)

// AwsElastiCacheClient is abstract interface of *elasticache.Client.
type AwsElastiCacheClient interface {
	DescribeCacheClusters(
		ctx context.Context,
		params *elasticache.DescribeCacheClustersInput,
		optFns ...func(*elasticache.Options),
	) (*elasticache.DescribeCacheClustersOutput, error)
	ListTagsForResource(
		ctx context.Context,
		params *elasticache.ListTagsForResourceInput,
		optFns ...func(*elasticache.Options),
	) (*elasticache.ListTagsForResourceOutput, error)
}

// AwsElastiCacheTagsMapper implements TagsMapper for AWS ElastiCache.
type AwsElastiCacheTagsMapper struct {
	cache  *goCache.Cache
	client AwsElastiCacheClient
}

// BuildAwsElastiCacheTagsMapper builds AwsElastiCacheTagsMapper from args.
func BuildAwsElastiCacheTagsMapper(cache *goCache.Cache, client AwsElastiCacheClient) AwsElastiCacheTagsMapper {
	return AwsElastiCacheTagsMapper{
		cache:  cache,
		client: client,
	}
}

// GetAwsElastiCacheClient returns AWS ElastiCache client.
func GetAwsElastiCacheClient(ctx context.Context) (*elasticache.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRetryer(func() aws.Retryer {
		return retry.AddWithMaxAttempts(retry.NewStandard(), 10)
	}))
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return elasticache.NewFromConfig(cfg), nil
}

// GetTagsMapping returns the latest tags mapping.
func (tm AwsElastiCacheTagsMapper) GetTagsMapping(ctx context.Context) (map[string]Tags, error) {
	if cv, found := tm.cache.Get(awsElastiCacheCacheKey); found {
		mapping := cv.(map[string]Tags)
		return mapping, nil
	}

	mapping := make(map[string]Tags)

	// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/elasticache#DescribeCacheClustersInput
	marker := aws.String("")

	for marker != nil {
		// Marker could not be empty string
		var input elasticache.DescribeCacheClustersInput
		if *marker == "" {
			input = elasticache.DescribeCacheClustersInput{}
		} else {
			input = elasticache.DescribeCacheClustersInput{Marker: marker}
		}

		output, err := tm.client.DescribeCacheClusters(ctx, &input)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		for i := 0; i < len(output.CacheClusters); i++ {
			cluster := output.CacheClusters[i]
			tagsInput := elasticache.ListTagsForResourceInput{ResourceName: cluster.ARN}
			tagsOutput, err := tm.client.ListTagsForResource(ctx, &tagsInput)
			if err != nil {
				return nil, fmt.Errorf("%w", err)
			}

			tags := make(Tags, len(tagsOutput.TagList))
			for j, tag := range tagsOutput.TagList {
				tags[j] = fmt.Sprintf("%s:%s", strings.ToLower(*tag.Key), strings.ToLower(*tag.Value))
			}
			mapping[*cluster.CacheClusterId] = tags
		}

		marker = output.Marker
	}

	tm.cache.Set(awsElastiCacheCacheKey, mapping, goCache.DefaultExpiration)
	return mapping, nil
}
