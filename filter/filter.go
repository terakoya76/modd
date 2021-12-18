package filter

import (
	"github.com/kelseyhightower/envconfig"

	"github.com/terakoya76/modd/aws"
	"github.com/terakoya76/modd/datadog"
)

// Filter is an interface to filter AWS resources which should be monitored.
type Filter interface {
	CheckScopeWithTags(scope datadog.Scope, tags aws.Tags) (included bool, excluded bool)
	CheckTagsWithTags(ddTags datadog.Tags, awsTags aws.Tags) bool
}

// BuildFilter build the proper filter object.
func BuildFilter(it datadog.IntegrationTarget) (Filter, error) {
	f := AwsFilter{
		AwsTagKey: "",
		DdTagKey:  "",
	}

	switch it { //nolint:gocritic
	case datadog.AwsRds:
		var c AwsRdsConfig
		err := envconfig.Process("aws_rds", &c)
		if err != nil {
			return nil, err
		}

		f = AwsFilter(c)
	}

	return f, nil
}
