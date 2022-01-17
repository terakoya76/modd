package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"sync"

	"github.com/terakoya76/modd/datadog"
	"github.com/terakoya76/modd/evaluator"
)

type monitorStatus struct {
	Name        string
	Unmonitored []string
}

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

	monitorStatuses, unsupported, err := checkUnmonitored(ctx, ddMonitorScopesMapping, ddMonitorTagsMapping)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to check monitor status: %v\n", err)
		os.Exit(1)
	}

	result := make(map[string]interface{})
	result["Monitors"] = monitorStatuses
	result["Unsupported"] = unsupported

	j, _ := json.Marshal(result)
	fmt.Fprintf(os.Stdout, "%s", j)
}

func checkUnmonitored(
	ctx context.Context,
	monitorScopesMapping datadog.MonitorScopesMapping,
	monitorTagsMapping datadog.MonitorTagsMapping,
) ([]monitorStatus, []string, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	unsupported := make([]string, 0)
	monitorStatuses := make([]monitorStatus, 0)
	for metric, scopes := range monitorScopesMapping {
		ddTags := monitorTagsMapping[metric]

		it := datadog.MetricToIntegrationTarget(metric)
		if it == datadog.UnknownIntegration {
			unsupported = append(unsupported, metric)
			continue
		}

		e, err := evaluator.BuildEvaluator(it)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get Evaluator object: %v", err)
		}

		wg.Add(1)
		go func(metric string, scopes []datadog.Scope) {
			defer wg.Done()

			unmonitored, err := e.Evaluate(ctx, scopes, ddTags)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to filter monitors: %v\n", err)
				return
			}

			mu.Lock()
			defer mu.Unlock()

			ms := monitorStatus{
				Name:        metric,
				Unmonitored: unmonitored,
			}

			monitorStatuses = append(monitorStatuses, ms)
		}(metric, scopes)
	}

	wg.Wait()

	sort.Slice(monitorStatuses, func(i, j int) bool {
		return monitorStatuses[i].Name < monitorStatuses[j].Name
	})
	for i := 0; i < len(monitorStatuses); i++ {
		sort.Strings(monitorStatuses[i].Unmonitored)
	}
	sort.Strings(unsupported)

	return monitorStatuses, unsupported, nil
}
