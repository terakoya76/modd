package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/terakoya76/modd/datadog"
	"github.com/terakoya76/modd/evaluator"
)

//nolint:funlen
func main() {
	ctx := datadog.GetDatadogContext()

	ddClient := datadog.GetDatadogClient()
	metadata, err := datadog.GetMetadata(ctx, ddClient)
	if err != nil {
		fmt.Fprintf(os.Stderr, "faield to get monitor metadata: %v\n", err)
		os.Exit(1)
	}

	monitors, err := datadog.ListMonitors(ctx, ddClient, metadata)
	if err != nil {
		fmt.Fprintf(os.Stderr, "faield to list monitors: %v\n", err)
		os.Exit(1)
	}

	ddMonitorTagsMapping, err := datadog.GetMonitorTagsMapping(monitors)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get monitor/tags mapping: %v\n", err)
		os.Exit(1)
	}

	ddMonitorScopesMapping, err := datadog.GetMonitorScopesMapping(monitors)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get monitor/scopes mapping: %v\n", err)
		os.Exit(1)
	}

	notMonitored := make(map[string][]string)
	for metric, scopes := range ddMonitorScopesMapping {
		ddTags := ddMonitorTagsMapping[metric]

		switch {
		case datadog.IsAwsRdsMetric(metric):
			e, err := evaluator.BuildEvaluator(datadog.AwsRds)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to get Evaluator object: %v\n", err)
				os.Exit(1)
			}

			ms, err := e.Evaluate(ctx, scopes, ddTags)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to filter monitors: %v\n", err)
				os.Exit(1)
			}

			notMonitored[metric] = ms
		case datadog.IsAwsElasticacheMetric(metric):
			e, err := evaluator.BuildEvaluator(datadog.AwsElasticache)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to get Evaluator object: %v\n", err)
				os.Exit(1)
			}

			ms, err := e.Evaluate(ctx, scopes, ddTags)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to filter monitors: %v\n", err)
				os.Exit(1)
			}

			notMonitored[metric] = ms
		case datadog.IsAwsOpenSearchServiceMetric(metric):
			e, err := evaluator.BuildEvaluator(datadog.AwsOpenSearchService)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to get Evaluator object: %v\n", err)
				os.Exit(1)
			}

			ms, err := e.Evaluate(ctx, scopes, ddTags)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to filter monitors: %v\n", err)
				os.Exit(1)
			}

			notMonitored[metric] = ms
		default:
			fmt.Printf("unsupported metrics: %s\n", metric)
		}
	}

	j, _ := json.MarshalIndent(notMonitored, "", "  ")
	fmt.Fprintf(os.Stdout, "%s\n", j)
}
