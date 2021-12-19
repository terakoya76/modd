package datadog

import (
	"strings"
)

// IntegrationTarget represents datadog integration service type.
type IntegrationTarget string

// AwsRds represents AWS RDS integration.
var AwsRds IntegrationTarget = "aws.rds"

// IsAwsRdsMetric determines if the given metric belongs to AWS RDS.
func IsAwsRdsMetric(metric string) bool {
	parts := strings.Split(metric, ".")
	return len(parts) >= 2 && parts[0] == "aws" && parts[1] == "rds"
}
