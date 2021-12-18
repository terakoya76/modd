package datadog

import (
	"context"
	"fmt"
	"os"
	"strings"

	dd "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
)

// Scope represents Datadog monitor scope
// cf. []string{"stage:production", "service:user"}
type Scope = []string

// MonitorScopesMapping represents a mapping of Datadog monitor ID and its scope
type MonitorScopesMapping = map[string][]Scope

// Tags represents Datadog tags
type Tags = []string

// MonitorTagsMapping represents a mapping of Datadog monitor ID and its tags
type MonitorTagsMapping = map[string]Tags

// GetDatadogContext returns Datadog authentication context
func GetDatadogContext() context.Context {
	return context.WithValue(
		context.Background(),
		dd.ContextAPIKeys,
		map[string]dd.APIKey{
			"apiKeyAuth": {
				Key: os.Getenv("DD_CLIENT_API_KEY"),
			},
			"appKeyAuth": {
				Key: os.Getenv("DD_CLIENT_APP_KEY"),
			},
		},
	)
}

// GetDatadogClient returns Datadog client
func GetDatadogClient() *dd.APIClient {
	configuration := dd.NewConfiguration()
	return dd.NewAPIClient(configuration)
}

// GetMetadata returns Datadog SearchMonitors Metadata
func GetMetadata(ctx context.Context, ddClient *dd.APIClient) (*dd.MonitorSearchResponseMetadata, error) {
	resp, _, err := ddClient.MonitorsApi.SearchMonitors(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	m := resp.GetMetadata()
	return &m, nil
}

// ListMonitors returns a list of Datadog monitors
func ListMonitors(ctx context.Context, ddClient *dd.APIClient, metadata *dd.MonitorSearchResponseMetadata) ([]dd.MonitorSearchResult, error) {
	monitors := make([]dd.MonitorSearchResult, 0, metadata.GetTotalCount())

	query := "type:integration"
	sort := "name,asc"
	perPage := int64(100)
	pages := int(metadata.GetTotalCount()/perPage + 1)

	for i := 0; i < pages; i++ {
		page := int64(i)
		optionalParams := dd.SearchMonitorsOptionalParameters{
			Query:   &query,
			Page:    &page,
			PerPage: &perPage,
			Sort:    &sort,
		}

		resp, _, err := ddClient.MonitorsApi.SearchMonitors(ctx, optionalParams)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		monitors = append(monitors, resp.GetMonitors()...)
	}

	return monitors, nil
}

// GetMonitorTagsMapping returns the latest MonitorTagsMapping
func GetMonitorTagsMapping(monitors []dd.MonitorSearchResult) (MonitorTagsMapping, error) {
	mapping := make(MonitorTagsMapping)

	for _, monitor := range monitors {
		tags := monitor.GetTags()

		for _, metric := range monitor.GetMetrics() {
			if _, ok := mapping[metric]; ok {
				mapping[metric] = append(mapping[metric], tags...)
			} else {
				mapping[metric] = tags
			}
		}
	}

	return mapping, nil
}

// GetMonitorScopesMapping returns the latest MonitorScopesMapping
func GetMonitorScopesMapping(monitors []dd.MonitorSearchResult) (MonitorScopesMapping, error) {
	mapping := make(MonitorScopesMapping)

	for _, monitor := range monitors {
		scopes := monitor.GetScopes()

		for _, metric := range monitor.GetMetrics() {
			if _, ok := mapping[metric]; ok {
				mapping[metric] = append(mapping[metric], scopes)
			} else {
				mapping[metric] = [][]string{scopes}
			}
		}
	}

	// remove duplicated scopes
	for metric, scopes := range mapping {
		mapping[metric] = makeUniq(scopes)
	}

	return mapping, nil
}

func makeUniq(arr [][]string) [][]string {
	// remove duplicated
	m := make(map[string]struct{})
	for _, elmt := range arr {
		m[strings.Join(elmt, ",")] = struct{}{}
	}

	uniq := make([][]string, 0, len(m))
	for joined := range m {
		if elmt := strings.Split(joined, ","); len(elmt) > 0 {
			uniq = append(uniq, elmt)
		}
	}

	return uniq
}
