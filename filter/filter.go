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
	switch it {
	case datadog.AwsElb:
		var c AwsElbConfig
		err := envconfig.Process("aws_elb", &c)
		if err != nil {
			return nil, err
		}

		f := AwsFilter(c)
		return f, nil
	case datadog.AwsAutoScalingGroup:
		var c AwsAutoScalingGroupConfig
		err := envconfig.Process("aws_autoscaling_group", &c)
		if err != nil {
			return nil, err
		}

		f := AwsFilter(c)
		return f, nil
	case datadog.AwsClb:
		var c AwsClbConfig
		err := envconfig.Process("aws_clb", &c)
		if err != nil {
			return nil, err
		}

		f := AwsFilter(c)
		return f, nil
	case datadog.AwsElasticache:
		var c AwsElasticacheConfig
		err := envconfig.Process("aws_elasticache", &c)
		if err != nil {
			return nil, err
		}

		f := AwsFilter(c)
		return f, nil
	case datadog.AwsKinesis:
		var c AwsKinesisConfig
		err := envconfig.Process("aws_kinesis", &c)
		if err != nil {
			return nil, err
		}

		f := AwsFilter(c)
		return f, nil
	case datadog.AwsOpenSearchService:
		var c AwsOpenSeardhServiceConfig
		err := envconfig.Process("aws_open_search_service", &c)
		if err != nil {
			return nil, err
		}

		f := AwsFilter(c)
		return f, nil
	case datadog.AwsRds:
		var c AwsRdsConfig
		err := envconfig.Process("aws_rds", &c)
		if err != nil {
			return nil, err
		}

		f := AwsFilter(c)
		return f, nil
	case datadog.AwsSqs:
		var c AwsSqsConfig
		err := envconfig.Process("aws_sqs", &c)
		if err != nil {
			return nil, err
		}

		f := AwsFilter(c)
		return f, nil
	default:
		return nil, fmt.Errorf("unsupported IntegrationTarget")
	}
}
