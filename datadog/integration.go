package datadog

import (
	"strings"
)

// IntegrationTarget represents datadog integration service type.
type IntegrationTarget string

var (
	// AwsMetricsPrefix represents Datadog AWS Integration metrics prefix.
	AwsMetricsPrefix = "aws"

	// UnknownIntegration represents unknonwn integration.
	UnknownIntegration IntegrationTarget = "unknown"
	// AwsAutoScalingGroup represents AWS AutoScalingGroup integration.
	AwsAutoScalingGroup IntegrationTarget = "aws.autoscaling"
	// AwsClb represents AWS CLB integration.
	AwsClb IntegrationTarget = "aws.elb"
	// AwsElasticache represents AWS Elasticache integration.
	AwsElasticache IntegrationTarget = "aws.elasticache"
	// AwsElb represents AWS ALB/NLB integration.
	AwsElb IntegrationTarget = "aws.applicationelb"
	// AwsKinesis represents AWS Kinesis integration.
	AwsKinesis IntegrationTarget = "aws.kinesis"
	// AwsOpenSeardhService represents AWS OpenSearch Service integration.
	AwsOpenSearchService IntegrationTarget = "aws.elasticsearchservice"
	// AwsRds represents AWS RDS integration.
	AwsRds IntegrationTarget = "aws.rds"
	// AwsSqs represents AWS SQS integration.
	AwsSqs IntegrationTarget = "aws.sqs"
)

// MetricToIntegrationTarget returns the IntegrationTarget to which the specified metric belongs.
//nolint:gocyclo
func MetricToIntegrationTarget(metric string) IntegrationTarget {
	parts := strings.Split(metric, ".")
	switch {
	case len(parts) >= 2 && parts[0] == AwsMetricsPrefix && parts[1] == "autoscaling":
		return AwsAutoScalingGroup
	case len(parts) >= 2 && parts[0] == AwsMetricsPrefix && parts[1] == "elb":
		return AwsClb
	case len(parts) >= 2 && parts[0] == AwsMetricsPrefix && parts[1] == "elasticache":
		return AwsElasticache
	case len(parts) >= 2 && parts[0] == AwsMetricsPrefix && parts[1] == "applicationelb":
		return AwsElb
	case len(parts) >= 2 && parts[0] == AwsMetricsPrefix && parts[1] == "kinesis":
		return AwsKinesis
	case len(parts) >= 2 && parts[0] == AwsMetricsPrefix && parts[1] == "es":
		return AwsOpenSearchService
	case len(parts) >= 2 && parts[0] == AwsMetricsPrefix && parts[1] == "rds":
		return AwsRds
	case len(parts) >= 2 && parts[0] == AwsMetricsPrefix && parts[1] == "sqs":
		return AwsSqs
	default:
		return UnknownIntegration
	}
}
