package evaluator

import (
	"context"
	"fmt"

	"github.com/terakoya76/modd/aws"
	"github.com/terakoya76/modd/datadog"
	"github.com/terakoya76/modd/filter"
)

type AwsRdsEvaluator struct{}

func (are AwsRdsEvaluator) GetMaaping(ctx context.Context) (map[string][]string, error) {
	rdsClient, err := aws.GetRdsClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get RdsClient: %w", err)
	}

	return aws.GetRdsTagsMapping(ctx, rdsClient)
}

// AwsRdsEvaluator returns a list of unmonitored AWS RDS identifiers.
func (are AwsRdsEvaluator) Evaluate(ctx context.Context, scopes []datadog.Scope, ddTags datadog.Tags) ([]string, error) {
	mapping, err := are.GetMaaping(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get AWS RDS Mapping: %w", err)
	}

	f, err := filter.BuildFilter(datadog.AwsRds)
	if err != nil {
		return nil, fmt.Errorf("failed to build Filter object: %w", err)
	}

	identifiers := aws.GetRdsIdentifiers(mapping)
	monitoredIdents := make([]string, 0, len(mapping))
	excludedIdents := make([]string, 0, len(mapping))

	for id, awsTags := range mapping {
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
