package datadog_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/terakoya76/modd/datadog"
)

func Test_MetricToIntegrationTarget(t *testing.T) {
	cases := []struct {
		name     string
		metric   string
		expected datadog.IntegrationTarget
	}{
		{
			name:     "when AWS API Gateway",
			metric:   "aws.apigateway.count",
			expected: datadog.AwsAPIGateway,
		},
		{
			name:     "when AWS AutoScalingGroup",
			metric:   "aws.autoscaling.group_desired_capacity",
			expected: datadog.AwsAutoScalingGroup,
		},
		{
			name:     "when AWS CLB",
			metric:   "aws.elb.latency",
			expected: datadog.AwsClb,
		},
		{
			name:     "when AWS DynamoDB",
			metric:   "aws.dynamodb.item_count",
			expected: datadog.AwsDynamoDB,
		},
		{
			name:     "when AWS ElastiCache",
			metric:   "aws.elasticache.cache_hits",
			expected: datadog.AwsElastiCache,
		},
		{
			name:     "when AWS ALB/NLB",
			metric:   "aws.applicationelb.healthy_host_count",
			expected: datadog.AwsElb,
		},
		{
			name:     "when AWS Firehose",
			metric:   "aws.firehose.incoming_records",
			expected: datadog.AwsFirehose,
		},
		{
			name:     "when AWS Kinesis",
			metric:   "aws.kinesis.get_records_latency",
			expected: datadog.AwsKinesis,
		},
		{
			name:     "when AWS OpenSearchService",
			metric:   "aws.es.elasticsearch_requests",
			expected: datadog.AwsOpenSearchService,
		},
		{
			name:     "when AWS RDS",
			metric:   "aws.rds.queries",
			expected: datadog.AwsRds,
		},
		{
			name:     "when AWS SNS",
			metric:   "aws.sns.number_of_notifications_failed",
			expected: datadog.AwsSns,
		},
		{
			name:     "when AWS StepFunction",
			metric:   "aws.states.executions_failed",
			expected: datadog.AwsStepFunction,
		},
		{
			name:     "when AWS SQS",
			metric:   "aws.sqs.number_of_messages_received",
			expected: datadog.AwsSqs,
		},
	}

	for _, c := range cases {
		actual := datadog.MetricToIntegrationTarget(c.metric)
		if !assert.Equal(t, c.expected, actual) {
			t.Errorf("case: %s is failed, expected: %v, actual: %v\n", c.name, c.expected, actual)
		}
	}
}
