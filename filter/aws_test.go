package filter_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/terakoya76/modd/datadog"
	"github.com/terakoya76/modd/filter"
	"github.com/terakoya76/modd/mapper"
)

func Test_CheckScopeWithTags(t *testing.T) {
	cases := []struct {
		name     string
		scope    datadog.Scope
		tags     mapper.Tags
		included bool
		excluded bool
	}{
		{
			name:     "when scope is wildcard without tags",
			scope:    []string{"*"},
			tags:     []string{},
			included: true,
			excluded: false,
		},
		{
			name:     "when scope is wildcard with a tag",
			scope:    []string{"*"},
			tags:     []string{"a:b"},
			included: true,
			excluded: false,
		},
		{
			name:     "when scope and tags are the same",
			scope:    []string{"a:b"},
			tags:     []string{"a:b"},
			included: true,
			excluded: false,
		},
		{
			name:     "when tags include scope",
			scope:    []string{"a:b"},
			tags:     []string{"a:b", "c:d"},
			included: true,
			excluded: false,
		},
		{
			name:     "when scope include tags",
			scope:    []string{"a:b", "c:d"},
			tags:     []string{"a:b"},
			included: false,
			excluded: false,
		},
		{
			name:     "when scope with inverted condition and tags are the same",
			scope:    []string{"!a:b"},
			tags:     []string{"a:b"},
			included: false,
			excluded: true,
		},
		{
			name:     "when scope with inverted condition includes tags",
			scope:    []string{"!a:b", "c:d"},
			tags:     []string{"a:b"},
			included: false,
			excluded: true,
		},
		{
			name:     "when tags include scope with inverted condition",
			scope:    []string{"!a:b"},
			tags:     []string{"a:b", "c:d"},
			included: false,
			excluded: true,
		},
	}

	for _, c := range cases {
		af := filter.AwsFilter{
			AwsTagKey: "",
			DdTagKey:  "",
		}

		included, excluded := af.CheckScopeWithTags(c.scope, c.tags)
		if !assert.Equal(t, c.included, included) {
			t.Errorf("case: %s is failed, expected: %t, actual: %t\n", c.name, c.included, included)
		}

		if !assert.Equal(t, c.excluded, excluded) {
			t.Errorf("case: %s is failed, expected: %t, actual: %t\n", c.name, c.excluded, excluded)
		}
	}
}

func Test_CheckTagsWithTags_Aws(t *testing.T) {
	cases := []struct {
		name     string
		filter   filter.Filter
		ddTags   datadog.Tags
		awsTags  mapper.Tags
		expected bool
	}{
		{
			name: "when filter holds no metadata",
			filter: filter.AwsFilter{
				AwsTagKey: "",
				DdTagKey:  "",
			},
			ddTags:   []string{},
			awsTags:  []string{},
			expected: true,
		},
		{
			name: "when filter holds no metadata even if tags exist",
			filter: filter.AwsFilter{
				AwsTagKey: "",
				DdTagKey:  "",
			},
			ddTags:   []string{"a:b"},
			awsTags:  []string{"c:d"},
			expected: true,
		},
		{
			name: "when filter holds only AWS metadata",
			filter: filter.AwsFilter{
				AwsTagKey: "a",
				DdTagKey:  "",
			},
			ddTags:   []string{},
			awsTags:  []string{},
			expected: true,
		},
		{
			name: "when filter holds only Datadog metadata",
			filter: filter.AwsFilter{
				AwsTagKey: "",
				DdTagKey:  "a",
			},
			ddTags:   []string{},
			awsTags:  []string{},
			expected: true,
		},
		{
			name: "when filter holds same key and tag's values are matched",
			filter: filter.AwsFilter{
				AwsTagKey: "a",
				DdTagKey:  "a",
			},
			ddTags:   []string{"a:b"},
			awsTags:  []string{"a:b"},
			expected: true,
		},
		{
			name: "when filter holds same key and tag's values are not matched",
			filter: filter.AwsFilter{
				AwsTagKey: "a",
				DdTagKey:  "a",
			},
			ddTags:   []string{"a:b"},
			awsTags:  []string{"a:c"},
			expected: false,
		},
		{
			name: "when filter holds different key and tag's values are matched",
			filter: filter.AwsFilter{
				AwsTagKey: "z",
				DdTagKey:  "a",
			},
			ddTags:   []string{"a:b"},
			awsTags:  []string{"z:b"},
			expected: true,
		},
		{
			name: "when filter holds different key and tag's values are not matched",
			filter: filter.AwsFilter{
				AwsTagKey: "z",
				DdTagKey:  "a",
			},
			ddTags:   []string{"a:b"},
			awsTags:  []string{"z:c"},
			expected: false,
		},
		{
			name: "when metadata and AWS tag are mismatched",
			filter: filter.AwsFilter{
				AwsTagKey: "a",
				DdTagKey:  "a",
			},
			ddTags:   []string{"a:b"},
			awsTags:  []string{"z:b"},
			expected: false,
		},
		{
			name: "when metadata and Datadog tag are mismatched",
			filter: filter.AwsFilter{
				AwsTagKey: "a",
				DdTagKey:  "a",
			},
			ddTags:   []string{"z:b"},
			awsTags:  []string{"a:b"},
			expected: false,
		},
	}

	for _, c := range cases {
		actual := c.filter.CheckTagsWithTags(c.ddTags, c.awsTags)
		if !assert.Equal(t, c.expected, actual) {
			t.Errorf("case: %s is failed, expected: %t, actual: %t\n", c.name, c.expected, actual)
		}
	}
}
