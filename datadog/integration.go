package datadog

import (
	"strings"
)

// IntegrationTarget represents datadog integration service type.
type IntegrationTarget string

const (
	// AwsMetricsPrefix represents Datadog AWS Integration metrics prefix.
	AwsMetricsPrefix = "aws"

	// AwsAPIGateway represents AWS API Gateway integration.
	AwsAPIGateway IntegrationTarget = "aws_apigateway"
	// AwsAutoScalingGroup represents AWS AutoScalingGroup integration.
	AwsAutoScalingGroup IntegrationTarget = "aws_autoscaling"
	// AwsClb represents AWS CLB integration.
	AwsClb IntegrationTarget = "aws_elb"
	// AwsDynamoDB represents AWS DynamoDB integration.
	AwsDynamoDB IntegrationTarget = "aws_dynamodb"
	// AwsElastiCache represents AWS ElastiCache integration.
	AwsElastiCache IntegrationTarget = "aws_elasticache"
	// AwsElb represents AWS ALB/NLB integration.
	AwsElb IntegrationTarget = "aws_applicationelb"
	// AwsFirehose represents AWS Firehose integration.
	AwsFirehose IntegrationTarget = "aws_firehose"
	// AwsKinesis represents AWS Kinesis integration.
	AwsKinesis IntegrationTarget = "aws_kinesis"
	// AwsLambda represents AWS Lambda integration.
	AwsLambda IntegrationTarget = "aws_lambda"
	// AwsOpenSearchService represents AWS OpenSearch Service integration.
	AwsOpenSearchService IntegrationTarget = "aws_elasticsearchservice"
	// AwsRds represents AWS RDS integration.
	AwsRds IntegrationTarget = "aws_rds"
	// AwsSns represents AWS SNS integration.
	AwsSns IntegrationTarget = "aws_sns"
	// AwsStepFunction represents AWS StepFunction integration.
	AwsStepFunction IntegrationTarget = "aws_states"
	// AwsSqs represents AWS SQS integration.
	AwsSqs IntegrationTarget = "aws_sqs"

	// UnknownIntegration represents unknonwn integration.
	UnknownIntegration IntegrationTarget = "unknown"
)

// MetricToIntegrationTarget returns the IntegrationTarget to which the specified metric belongs.
//nolint:gocyclo
func MetricToIntegrationTarget(metric string) IntegrationTarget {
	parts := strings.Split(metric, ".")
	switch {
	case len(parts) >= 2 && parts[0] == AwsMetricsPrefix && parts[1] == "apigateway":
		return AwsAPIGateway
	case len(parts) >= 2 && parts[0] == AwsMetricsPrefix && parts[1] == "autoscaling":
		return AwsAutoScalingGroup
	case len(parts) >= 2 && parts[0] == AwsMetricsPrefix && parts[1] == "elb":
		return AwsClb
	case len(parts) >= 2 && parts[0] == AwsMetricsPrefix && parts[1] == "dynamodb":
		return AwsDynamoDB
	case len(parts) >= 2 && parts[0] == AwsMetricsPrefix && parts[1] == "elasticache":
		return AwsElastiCache
	case len(parts) >= 2 && parts[0] == AwsMetricsPrefix && parts[1] == "applicationelb":
		return AwsElb
	case len(parts) >= 2 && parts[0] == AwsMetricsPrefix && parts[1] == "firehose":
		return AwsFirehose
	case len(parts) >= 2 && parts[0] == AwsMetricsPrefix && parts[1] == "kinesis":
		return AwsKinesis
	case len(parts) >= 2 && parts[0] == AwsMetricsPrefix && parts[1] == "lambda":
		return AwsLambda
	case len(parts) >= 2 && parts[0] == AwsMetricsPrefix && parts[1] == "es":
		return AwsOpenSearchService
	case len(parts) >= 2 && parts[0] == AwsMetricsPrefix && parts[1] == "rds":
		return AwsRds
	case len(parts) >= 2 && parts[0] == AwsMetricsPrefix && parts[1] == "sns":
		return AwsSns
	case len(parts) >= 2 && parts[0] == AwsMetricsPrefix && parts[1] == "states":
		return AwsStepFunction
	case len(parts) >= 2 && parts[0] == AwsMetricsPrefix && parts[1] == "sqs":
		return AwsSqs
	default:
		return UnknownIntegration
	}
}
