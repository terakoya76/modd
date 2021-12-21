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
)

// AwsElasticacheTagsMapper implements TagsMapper for AWS Elasticache.
type AwsElasticacheTagsMapper struct {
	cache  *cache.Cache
	client *elasticache.Client
}

// GetAwsElasticacheClient returns AWS Elasticache client.
func GetAwsElasticacheClient(ctx context.Context) (*elasticache.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRetryer(func() aws.Retryer {
		return retry.AddWithMaxAttempts(retry.NewStandard(), 10)
	}))
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return elasticache.NewFromConfig(cfg), nil
}

// GetTagsMapping returns the latest tags mapping.
func (aetm AwsElasticacheTagsMapper) GetTagsMapping(ctx context.Context) (map[string]Tags, error) {
	if cv, found := aetm.cache.Get(awsElasticacheCache); found {
		mapping := cv.(map[string]Tags)
		return mapping, nil
	}

	mapping := make(map[string]Tags)

	initMarker := ""
	marker := &initMarker

	for marker != nil {
		input := elasticache.DescribeCacheClustersInput{Marker: marker}
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

	aetm.cache.Set(awsElasticacheCache, mapping, cache.DefaultExpiration)
	return mapping, nil
}
