package filter

import (
	"strings"

	"github.com/terakoya76/modd/aws"
	"github.com/terakoya76/modd/datadog"
)

// AwsFilter implements Filter interface
// it holds the metadata to filter AWS resources which should be monitored
type AwsFilter struct {
	AwsTagKey string
	DdTagKey  string
}

// CheckScopeWithTags evaluates Datadog scope and AWS resources
func (af AwsFilter) CheckScopeWithTags(scope datadog.Scope, tags aws.Tags) (included bool, excluded bool) {
	wildcard := false
	matchers := make([]string, 0, len(scope))
	inverted := make([]string, 0, len(scope))
	for _, matcher := range scope {
		if matcher[0] == '!' {
			inverted = append(inverted, matcher[1:])
		} else if matcher == "*" {
			wildcard = true
		} else {
			matchers = append(matchers, matcher)
		}
	}

	if len(Intersect(inverted, tags)) > 0 {
		return false, true
	}

	if wildcard {
		return true, false
	}

	if len(Difference(matchers, tags)) == 0 {
		return true, false
	}

	return false, false
}

// CheckTagsWithTags evaluates a Datadog/AWS tag matcher
func (af AwsFilter) CheckTagsWithTags(ddTags datadog.Tags, awsTags aws.Tags) bool {
	if af.AwsTagKey == "" || af.DdTagKey == "" {
		return true
	}

	for _, dt := range ddTags {
		dparts := strings.Split(dt, ":")
		dk, dv := dparts[0], dparts[1]
		if dk != af.DdTagKey {
			continue
		}

		for _, at := range awsTags {
			aparts := strings.Split(at, ":")
			ak, av := aparts[0], aparts[1]
			if ak != af.AwsTagKey {
				continue
			}

			if dv == av {
				return true
			}
		}
	}

	return false
}
