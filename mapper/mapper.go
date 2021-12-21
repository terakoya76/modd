package mapper

import (
	"context"
	"fmt"
	"time"

	"github.com/patrickmn/go-cache"

	"github.com/terakoya76/modd/datadog"
)

// Tags represents resource tags.
type Tags = []string

// TagsMapper is an interface to fetch resources and map their ids and tags.
type TagsMapper interface {
	GetTagsMapping(ctx context.Context) (map[string]Tags, error)
}

// BuildTagsMapper build the proper TagsMapper implementation.
func BuildTagsMapper(it datadog.IntegrationTarget) (TagsMapper, error) {
	c := cache.New(10*time.Minute, 1*time.Minute)

	switch it {
	case datadog.AwsAutoScalingGroup:
		client, err := GetAwsAutoScalingGroupClient(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		m := AwsAutoScalingGroupTagsMapper{
			cache:  c,
			client: client,
		}

		return m, nil
	case datadog.AwsElasticache:
		client, err := GetAwsElasticacheClient(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		m := AwsElasticacheTagsMapper{
			cache:  c,
			client: client,
		}

		return m, nil
	case datadog.AwsKinesis:
		client, err := GetAwsKinesisClient(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		m := AwsKinesisTagsMapper{
			cache:  c,
			client: client,
		}

		return m, nil
	case datadog.AwsOpenSearchService:
		client, err := GetAwsOpenSearchServiceClient(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		m := AwsOpenSearchServiceTagsMapper{
			cache:  c,
			client: client,
		}

		return m, nil
	case datadog.AwsRds:
		client, err := GetAwsRdsClient(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		m := AwsRdsTagsMapper{
			cache:  c,
			client: client,
		}

		return m, nil
	case datadog.AwsSqs:
		client, err := GetAwsSqsClient(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		m := AwsSqsTagsMapper{
			cache:  c,
			client: client,
		}

		return m, nil
	default:
		return nil, fmt.Errorf("unsupported IntegrationTarget")
	}
}
