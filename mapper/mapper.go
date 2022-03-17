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
//nolint:funlen,gocyclo
func BuildTagsMapper(it datadog.IntegrationTarget) (TagsMapper, error) {
	c := cache.New(60*time.Minute, 10*time.Minute)

	switch it {
	case datadog.AwsAPIGateway:
		client, err := GetAwsAPIGatewayClient(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		m := BuildAwsAPIGatewayTagsMapper(c, client)
		return m, nil

	case datadog.AwsAutoScalingGroup:
		client, err := GetAwsAutoScalingGroupClient(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		m := BuildAwsAutoScalingGroupTagsMapper(c, client)
		return m, nil

	case datadog.AwsClb:
		client, err := GetAwsClbClient(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		m := AwsClbTagsMapper{
			cache:  c,
			client: client,
		}

		return m, nil

	case datadog.AwsDynamoDB:
		client, err := GetAwsDynamoDBClient(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		m := BuildAwsDynamoDBTagsMapper(c, client)
		return m, nil

	case datadog.AwsElastiCache:
		client, err := GetAwsElastiCacheClient(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		m := BuildAwsElastiCacheTagsMapper(c, client)
		return m, nil

	case datadog.AwsElb:
		client, err := GetAwsElbClient(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		m := AwsElbTagsMapper{
			cache:  c,
			client: client,
		}

		return m, nil

	case datadog.AwsFirehose:
		client, err := GetAwsFirehoseClient(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		m := BuildAwsFirehoseTagsMapper(c, client)
		return m, nil

	case datadog.AwsKinesis:
		client, err := GetAwsKinesisClient(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		m := BuildAwsKinesisTagsMapper(c, client)
		return m, nil

	case datadog.AwsOpenSearchService:
		client, err := GetAwsOpenSearchServiceClient(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		m := BuildAwsOpenSearchServiceTagsMapper(c, client)
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

	case datadog.AwsSns:
		client, err := GetAwsSnsClient(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		m := BuildAwsSnsTagsMapper(c, client)
		return m, nil

	case datadog.AwsStepFunction:
		client, err := GetAwsStepFunctionClient(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		m := BuildAwsStepFunctionTagsMapper(c, client)
		return m, nil

	case datadog.AwsSqs:
		client, err := GetAwsSqsClient(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		m := BuildAwsSqsTagsMapper(c, client)
		return m, nil

	case datadog.UnknownIntegration:
		return nil, fmt.Errorf("unsupported IntegrationTarget")
	default:
		return nil, fmt.Errorf("unsupported IntegrationTarget")
	}
}
