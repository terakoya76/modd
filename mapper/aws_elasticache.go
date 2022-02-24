package mapper

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	"github.com/patrickmn/go-cache"

	"github.com/terakoya76/modd/datadog"
)

const awsElastiCacheCacheKey string = string(datadog.AwsElastiCache)

// AwsElastiCacheTagsMapper implements TagsMapper for AWS ElastiCache.
type AwsElastiCacheTagsMapper struct {
	cache  *cache.Cache
	client *elasticache.Client
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
func (aetm AwsElastiCacheTagsMapper) GetTagsMapping(ctx context.Context) (map[string]Tags, error) {
	if cv, found := aetm.cache.Get(awsElastiCacheCacheKey); found {
		mapping := cv.(map[string]Tags)
		return mapping, nil
	}

	mapping := make(map[string]Tags)

	// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/elasticache#DescribeCacheClustersInput
	initMarker := ""
	marker := &initMarker

	for marker != nil {
		// Marker could not be empty string
		var input elasticache.DescribeCacheClustersInput
		if *marker == "" {
			input = elasticache.DescribeCacheClustersInput{}
		} else {
			input = elasticache.DescribeCacheClustersInput{Marker: marker}
		}

		output, err := aetm.client.DescribeCacheClusters(ctx, &input)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		for i := 0; i < len(output.CacheClusters); i++ {
			cluster := output.CacheClusters[i]
			tagsInput := elasticache.ListTagsForResourceInput{ResourceName: cluster.ARN}
			tagsOutput, err := aetm.client.ListTagsForResource(ctx, &tagsInput)
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

	aetm.cache.Set(awsElastiCacheCacheKey, mapping, cache.DefaultExpiration)
	return mapping, nil
}
