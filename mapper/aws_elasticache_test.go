package mapper_test

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	"github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/aws/smithy-go/middleware"
	goCache "github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"

	"github.com/terakoya76/modd/mapper"
)

// dummyAwsElastiCacheClient implements AwsElastiCacheClient interface for faking AWS API.
type dummyAwsElastiCacheClient struct{}

// DescribeCacheClusters implements AwsElastiCacheClient for dummyAwsElastiCacheClient.
func (c *dummyAwsElastiCacheClient) DescribeCacheClusters(
	_ context.Context,
	params *elasticache.DescribeCacheClustersInput,
	_ ...func(*elasticache.Options),
) (*elasticache.DescribeCacheClustersOutput, error) {
	var output elasticache.DescribeCacheClustersOutput

	if params.Marker != nil {
		output = elasticache.DescribeCacheClustersOutput{
			CacheClusters: []types.CacheCluster{
				{
					ARN:            aws.String("cache10"),
					CacheClusterId: aws.String("cache10"),
				},
				{
					ARN:            aws.String("cache20"),
					CacheClusterId: aws.String("cache20"),
				},
			},
			Marker:         nil,
			ResultMetadata: middleware.Metadata{},
		}
	} else {
		output = elasticache.DescribeCacheClustersOutput{
			CacheClusters: []types.CacheCluster{
				{
					ARN:            aws.String("cache1"),
					CacheClusterId: aws.String("cache1"),
				},
				{
					ARN:            aws.String("cache2"),
					CacheClusterId: aws.String("cache2"),
				},
			},
			Marker:         aws.String("next token"),
			ResultMetadata: middleware.Metadata{},
		}
	}

	return &output, nil
}

// ListTagsForResource implements AwsElastiCacheClient for dummyAwsElastiCacheClient.
func (c *dummyAwsElastiCacheClient) ListTagsForResource(
	_ context.Context,
	_ *elasticache.ListTagsForResourceInput,
	_ ...func(*elasticache.Options),
) (*elasticache.ListTagsForResourceOutput, error) {
	output := elasticache.ListTagsForResourceOutput{
		TagList: []types.Tag{
			{
				Key:   aws.String("key1"),
				Value: aws.String("val1"),
			},
			{
				Key:   aws.String("key2"),
				Value: aws.String("val2"),
			},
			{
				Key:   aws.String("key10"),
				Value: aws.String("val10"),
			},
			{
				Key:   aws.String("key20"),
				Value: aws.String("val20"),
			},
		},
		ResultMetadata: middleware.Metadata{},
	}

	return &output, nil
}

func Test_AwsElastiCache_GetTagsMapping(t *testing.T) {
	cache := goCache.New(60*time.Minute, 10*time.Minute)

	cases := []struct {
		name     string
		expected map[string]mapper.Tags
		err      error
	}{
		{
			name: "fake test",
			expected: map[string]mapper.Tags{
				"cache1":  []string{"key1:val1", "key2:val2", "key10:val10", "key20:val20"},
				"cache2":  []string{"key1:val1", "key2:val2", "key10:val10", "key20:val20"},
				"cache10": []string{"key1:val1", "key2:val2", "key10:val10", "key20:val20"},
				"cache20": []string{"key1:val1", "key2:val2", "key10:val10", "key20:val20"},
			},
			err: nil,
		},
	}

	for _, c := range cases {
		client := dummyAwsElastiCacheClient{}
		m := mapper.BuildAwsElastiCacheTagsMapper(cache, &client)
		actual, err := m.GetTagsMapping(context.TODO())
		if !assert.Equal(t, c.err, err) {
			t.Errorf("case: %s is failed, expected: %+v, actual: %+v\n", c.name, c.err, err)
		}
		if !assert.Equal(t, c.expected, actual) {
			t.Errorf("case: %s is failed, expected: %+v, actual: %+v\n", c.name, c.expected, actual)
		}
	}
}
