package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/terakoya76/modd/datadog"
	"github.com/terakoya76/modd/evaluator"
)

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

	var wg sync.WaitGroup
	var mu sync.Mutex

	notMonitored := make(map[string][]string)
	for metric, scopes := range ddMonitorScopesMapping {
		ddTags := ddMonitorTagsMapping[metric]

		it := datadog.MetricToIntegrationTarget(metric)
		if it == datadog.UnknownIntegration {
			fmt.Printf("unsupported metrics: %s\n", metric)
			continue
		}

		e, err := evaluator.BuildEvaluator(it)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to get Evaluator object: %v\n", err)
			os.Exit(1)
		}

		wg.Add(1)
		go func(metric string, scopes []datadog.Scope) {
			defer wg.Done()

			ms, err := e.Evaluate(ctx, scopes, ddTags)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to filter monitors: %v\n", err)
				return
			}

			mu.Lock()
			defer mu.Unlock()

			notMonitored[metric] = ms
		}(metric, scopes)
	}

	wg.Wait()

	j, _ := json.MarshalIndent(notMonitored, "", "  ")
	fmt.Fprintf(os.Stdout, "%s\n", j)
}
