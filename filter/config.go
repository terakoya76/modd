package filter

// AwsAPIGatewayConfig holds metadata for AwsFilter for AWS API Gateway.
type AwsAPIGatewayConfig struct {
	AwsTagKey string `envconfig:"aws_tag_key" default:""`
	DdTagKey  string `envconfig:"datadog_tag_key" default:""`
}

// AwsAutoScalingGroupConfig holds metadata for AwsFilter for AWS AutoScalingGroup.
type AwsAutoScalingGroupConfig struct {
	AwsTagKey string `envconfig:"aws_tag_key" default:""`
	DdTagKey  string `envconfig:"datadog_tag_key" default:""`
}

// AwsClbConfig holds metadata for AwsFilter for AWS CLB.
type AwsClbConfig struct {
	AwsTagKey string `envconfig:"aws_tag_key" default:""`
	DdTagKey  string `envconfig:"datadog_tag_key" default:""`
}

// AwsDynamoDBConfig holds metadata for AwsFilter for AWS DynamoDB.
type AwsDynamoDBConfig struct {
	AwsTagKey string `envconfig:"aws_tag_key" default:""`
	DdTagKey  string `envconfig:"datadog_tag_key" default:""`
}

// AwsElastiCacheConfig holds metadata for AwsFilter for AWS ElastiCache.
type AwsElastiCacheConfig struct {
	AwsTagKey string `envconfig:"aws_tag_key" default:""`
	DdTagKey  string `envconfig:"datadog_tag_key" default:""`
}

// AwsElbConfig holds metadata for AwsFilter for AWS ALB.
type AwsElbConfig struct {
	AwsTagKey string `envconfig:"aws_tag_key" default:""`
	DdTagKey  string `envconfig:"datadog_tag_key" default:""`
}

// AwsFirehoseConfig holds metadata for AwsFilter for AWS Firehose.
type AwsFirehoseConfig struct {
	AwsTagKey string `envconfig:"aws_tag_key" default:""`
	DdTagKey  string `envconfig:"datadog_tag_key" default:""`
}

// AwsKinesisConfig holds metadata for AwsFilter for AWS Kinesis.
type AwsKinesisConfig struct {
	AwsTagKey string `envconfig:"aws_tag_key" default:""`
	DdTagKey  string `envconfig:"datadog_tag_key" default:""`
}

// AwsLambdaConfig holds metadata for AwsFilter for AWS Lambda.
type AwsLambdaConfig struct {
	AwsTagKey string `envconfig:"aws_tag_key" default:""`
	DdTagKey  string `envconfig:"datadog_tag_key" default:""`
}

// AwsOpenSeardhServiceConfig holds metadata for AwsFilter for AWS OpenSearch Service.
type AwsOpenSeardhServiceConfig struct {
	AwsTagKey string `envconfig:"aws_tag_key" default:""`
	DdTagKey  string `envconfig:"datadog_tag_key" default:""`
}

// AwsRdsConfig holds metadata for AwsFilter for AWS RDS.
type AwsRdsConfig struct {
	AwsTagKey string `envconfig:"aws_tag_key" default:""`
	DdTagKey  string `envconfig:"datadog_tag_key" default:""`
}

// AwsSnsConfig holds metadata for AwsFilter for AWS SNS.
type AwsSnsConfig struct {
	AwsTagKey string `envconfig:"aws_tag_key" default:""`
	DdTagKey  string `envconfig:"datadog_tag_key" default:""`
}

// AwsStepFunctionConfig holds metadata for AwsFilter for AWS StepFunction.
type AwsStepFunctionConfig struct {
	AwsTagKey string `envconfig:"aws_tag_key" default:""`
	DdTagKey  string `envconfig:"datadog_tag_key" default:""`
}

// AwsSqsConfig holds metadata for AwsFilter for AWS SQS.
type AwsSqsConfig struct {
	AwsTagKey string `envconfig:"aws_tag_key" default:""`
	DdTagKey  string `envconfig:"datadog_tag_key" default:""`
}
