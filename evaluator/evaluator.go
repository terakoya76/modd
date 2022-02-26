package evaluator

import (
	"context"
	"fmt"

	"golang.org/x/sync/singleflight"

	"github.com/terakoya76/modd/datadog"
	"github.com/terakoya76/modd/filter"
	"github.com/terakoya76/modd/mapper"
)

var group singleflight.Group

// Evaluator gets target resources and tags via TagsMapper and filter them via Filter.
type Evaluator struct {
	it        datadog.IntegrationTarget
	filter    filter.Filter
	tagMapper mapper.TagsMapper
}

// BuildEvaluator build the proper Evaluator implementation.
func BuildEvaluator(it datadog.IntegrationTarget) (Evaluator, error) {
	f, err := filter.BuildFilter(it)
	if err != nil {
		return Evaluator{}, fmt.Errorf("failed to get Filter object")
	}

	m, err := mapper.BuildTagsMapper(it)
	if err != nil {
		return Evaluator{}, fmt.Errorf("failed to get TagsMapper object")
	}

	e := Evaluator{
		it:        it,
		filter:    f,
		tagMapper: m,
	}

	return e, nil
}

// Evaluate returns a list of unmonitored resource identifiers.
func (e Evaluator) Evaluate(ctx context.Context, scopes []datadog.Scope, ddTags datadog.Tags) ([]string, error) {
	name := string(e.it)
	v, err, _ := group.Do(name, func() (interface{}, error) {
		return e.tagMapper.GetTagsMapping(ctx)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get resource tags mapping: %w", err)
	}

	mapping, ok := v.(map[string][]string)
	if !ok {
		return nil, fmt.Errorf("failed type assertion: %w", err)
	}

	identifiers := GetIdentifiersFromMaaping(mapping)
	monitoredIdents := make([]string, 0, len(mapping))
	excludedIdents := make([]string, 0, len(mapping))

	for id, resourceTags := range mapping {
		for _, scope := range scopes {
			monitored, excluded := e.filter.CheckScopeWithTags(scope, resourceTags)
			if monitored {
				monitoredIdents = append(monitoredIdents, id)
			}
			if excluded {
				excludedIdents = append(excludedIdents, id)
			}
		}

		if !e.filter.CheckTagsWithTags(ddTags, resourceTags) {
			excludedIdents = append(excludedIdents, id)
		}
	}

	return filter.Difference(filter.Difference(identifiers, monitoredIdents), excludedIdents), nil
}

// GetIdentifiersFromMaaping returns a list of identifiers from mapping keys.
func GetIdentifiersFromMaaping(mapping map[string][]string) []string {
	keys := make([]string, len(mapping))
	i := 0
	for k := range mapping {
		keys[i] = k
		i++
	}

	return keys
}
