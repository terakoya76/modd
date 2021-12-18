package filter

import ()

// AwsRdsConfig holds metadata for AwsFilter for AWS RDS
type AwsRdsConfig struct {
	AwsTagKey string `envconfig:"aws_tag_key" default:""`
	DdTagKey  string `envconfig:"datadog_tag_key" default:""`
}
