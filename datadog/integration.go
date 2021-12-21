package datadog

import (
	"strings"
)

// IntegrationTarget represents datadog integration service type.
type IntegrationTarget string

var (
	// AwsMetricsPrefix represents Datadog AWS Integration metrics prefix.
	AwsMetricsPrefix = "aws"

	// AwsAutoScalingGroup represents AWS AutoScalingGroup integration.
	AwsAutoScalingGroup IntegrationTarget = "aws.autoscaling"
	// AwsElasticache represents AWS Elasticache integration.
	AwsElasticache IntegrationTarget = "aws.elasticache"
	// AwsKinesis represents AWS Kinesis integration.
	AwsKinesis IntegrationTarget = "aws.kinesis"
	// AwsOpenSeardhService represents AWS OpenSearch Service integration.
	AwsOpenSearchService IntegrationTarget = "aws.elasticsearchservice"
	// AwsRds represents AWS RDS integration.
	AwsRds IntegrationTarget = "aws.rds"
	// AwsSqs represents AWS SQS integration.
	AwsSqs IntegrationTarget = "aws.sqs"
)

// IsAwsAutoScalingGroupMetric determines if the given metric belongs to AWS AutoScalingGroup.
func IsAwsAutoScalingGroupMetric(metric string) bool {
	parts := strings.Split(metric, ".")
	return len(parts) >= 2 && parts[0] == AwsMetricsPrefix && parts[1] == "autoscaling"
}

// IsAwsElasticacheMetric determines if the given metric belongs to AWS Elasticache.
func IsAwsElasticacheMetric(metric string) bool {
	parts := strings.Split(metric, ".")
	return len(parts) >= 2 && parts[0] == AwsMetricsPrefix && parts[1] == "elasticache"
}

// IsAwsKinesisMetric determines if the given metric belongs to AWS Kinesis.
func IsAwsKinesisMetric(metric string) bool {
	parts := strings.Split(metric, ".")
	return len(parts) >= 2 && parts[0] == AwsMetricsPrefix && parts[1] == "kinesis"
}

// IsAwsOpenSearchServiceMetric determines if the given metric belongs to AWS OpenSearch Service.
func IsAwsOpenSearchServiceMetric(metric string) bool {
	parts := strings.Split(metric, ".")
	return len(parts) >= 2 && parts[0] == AwsMetricsPrefix && parts[1] == "es"
}

// IsAwsRdsMetric determines if the given metric belongs to AWS RDS.
func IsAwsRdsMetric(metric string) bool {
	parts := strings.Split(metric, ".")
	return len(parts) >= 2 && parts[0] == AwsMetricsPrefix && parts[1] == "rds"
}

// IsAwsSqsMetric determines if the given metric belongs to AWS SQS.
func IsAwsSqsMetric(metric string) bool {
	parts := strings.Split(metric, ".")
	return len(parts) >= 2 && parts[0] == AwsMetricsPrefix && parts[1] == "sqs"
}
