package filter

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"

	"github.com/terakoya76/modd/datadog"
	"github.com/terakoya76/modd/mapper"
)

// Filter is an interface to filter AWS resources which should be monitored.
type Filter interface {
	CheckScopeWithTags(scope datadog.Scope, tags mapper.Tags) (included bool, excluded bool)
	CheckTagsWithTags(ddTags datadog.Tags, resourceTags mapper.Tags) bool
}

// BuildFilter build the proper Filter implementation.
//nolint:funlen,gocyclo
func BuildFilter(it datadog.IntegrationTarget) (Filter, error) {
	envPrefix := string(it)

	switch it {
	case datadog.AwsElb:
		var c AwsElbConfig
		err := envconfig.Process(envPrefix, &c)
		if err != nil {
			return nil, err
		}

		f := AwsFilter(c)
		return f, nil
	case datadog.AwsAutoScalingGroup:
		var c AwsAutoScalingGroupConfig
		err := envconfig.Process(envPrefix, &c)
		if err != nil {
			return nil, err
		}

		f := AwsFilter(c)
		return f, nil
	case datadog.AwsClb:
		var c AwsClbConfig
		err := envconfig.Process(envPrefix, &c)
		if err != nil {
			return nil, err
		}

		f := AwsFilter(c)
		return f, nil
	case datadog.AwsDynamoDB:
		var c AwsDynamoDBConfig
		err := envconfig.Process(envPrefix, &c)
		if err != nil {
			return nil, err
		}

		f := AwsFilter(c)
		return f, nil
	case datadog.AwsElastiCache:
		var c AwsElastiCacheConfig
		err := envconfig.Process(envPrefix, &c)
		if err != nil {
			return nil, err
		}

		f := AwsFilter(c)
		return f, nil
	case datadog.AwsKinesis:
		var c AwsKinesisConfig
		err := envconfig.Process(envPrefix, &c)
		if err != nil {
			return nil, err
		}

		f := AwsFilter(c)
		return f, nil
	case datadog.AwsOpenSearchService:
		var c AwsOpenSeardhServiceConfig
		err := envconfig.Process(envPrefix, &c)
		if err != nil {
			return nil, err
		}

		f := AwsFilter(c)
		return f, nil
	case datadog.AwsRds:
		var c AwsRdsConfig
		err := envconfig.Process(envPrefix, &c)
		if err != nil {
			return nil, err
		}

		f := AwsFilter(c)
		return f, nil
	case datadog.AwsStepFunction:
		var c AwsStepFunctionConfig
		err := envconfig.Process(envPrefix, &c)
		if err != nil {
			return nil, err
		}

		f := AwsFilter(c)
		return f, nil
	case datadog.AwsSqs:
		var c AwsSqsConfig
		err := envconfig.Process(envPrefix, &c)
		if err != nil {
			return nil, err
		}

		f := AwsFilter(c)
		return f, nil
	case datadog.UnknownIntegration:
		return nil, fmt.Errorf("unsupported IntegrationTarget")
	default:
		return nil, fmt.Errorf("unsupported IntegrationTarget")
	}
}
