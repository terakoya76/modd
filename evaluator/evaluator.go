package evaluator

import (
	"context"

	"github.com/terakoya76/modd/datadog"
)

// Evaluator is an interface to get target resources and filter by Datadog scopes and tags matchers.
type Evaluator interface {
	GetMaaping() map[string][]string
	Evaluate(ctx context.Context, scopes []datadog.Scope, ddTags datadog.Tags) ([]string, error)
}
