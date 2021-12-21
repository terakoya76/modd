package filter

// AwsRdsConfig holds metadata for AwsFilter for AWS RDS.
type AwsRdsConfig struct {
	AwsTagKey string `envconfig:"aws_tag_key" default:""`
	DdTagKey  string `envconfig:"datadog_tag_key" default:""`
}

// AwsElasticacheConfig holds metadata for AwsFilter for AWS Elasticache.
type AwsElasticacheConfig struct {
	AwsTagKey string `envconfig:"aws_tag_key" default:""`
	DdTagKey  string `envconfig:"datadog_tag_key" default:""`
}

// AwsOpenSeardhServiceConfig holds metadata for AwsFilter for AWS OpenSearch Service.
type AwsOpenSeardhServiceConfig struct {
	AwsTagKey string `envconfig:"aws_tag_key" default:""`
	DdTagKey  string `envconfig:"datadog_tag_key" default:""`
}

// AwsSqsConfig holds metadata for AwsFilter for AWS SQS.
type AwsSqsConfig struct {
	AwsTagKey string `envconfig:"aws_tag_key" default:""`
	DdTagKey  string `envconfig:"datadog_tag_key" default:""`
}

// AwsKinesisConfig holds metadata for AwsFilter for AWS Kinesis.
type AwsKinesisConfig struct {
	AwsTagKey string `envconfig:"aws_tag_key" default:""`
	DdTagKey  string `envconfig:"datadog_tag_key" default:""`
}
