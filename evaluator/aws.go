package evaluator

import (
	"context"
	"fmt"

	"github.com/terakoya76/modd/aws"
	"github.com/terakoya76/modd/datadog"
	"github.com/terakoya76/modd/filter"
)

// AwsRdsEvaluate returns a list of unmonitored AWS RDS identifiers
func AwsRdsEvaluate(ctx context.Context, scopes []datadog.Scope, ddTags datadog.Tags) ([]string, error) {
	rdsClient, err := aws.GetRdsClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get RdsClient: %w", err)
	}

	awsRdsMapping, err := aws.GetRdsTagsMapping(ctx, rdsClient)
	if err != nil {
		return nil, fmt.Errorf("failed to get rds mapping: %w", err)
	}

	identifiers := aws.GetRdsIdentifiers(awsRdsMapping)
	monitoredIdents := make([]string, 0, len(awsRdsMapping))
	excludedIdents := make([]string, 0, len(awsRdsMapping))

	f, err := filter.BuildFilter(datadog.AwsRds)
	if err != nil {
		return nil, fmt.Errorf("failed to build Filter object: %w", err)
	}

	for id, awsTags := range awsRdsMapping {
		for _, scope := range scopes {
			monitored, excluded := f.CheckScopeWithTags(scope, awsTags)
			if monitored {
				monitoredIdents = append(monitoredIdents, id)
			}
			if excluded {
				excludedIdents = append(excludedIdents, id)
			}
		}

		if !f.CheckTagsWithTags(ddTags, awsTags) {
			excludedIdents = append(excludedIdents, id)
		}
	}

	return filter.Difference(filter.Difference(identifiers, monitoredIdents), excludedIdents), nil
}
