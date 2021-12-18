package filter_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/terakoya76/modd/filter"
)

func TestIntersect(t *testing.T) {
	cases := []struct {
		name     string
		arg1     []string
		arg2     []string
		expected []string
	}{
		{
			name:     "when arg1/arg2 includes the mutual item",
			arg1:     []string{"a", "b"},
			arg2:     []string{"b", "c"},
			expected: []string{"b"},
		},
		{
			name:     "when arg1/arg2 are the same set",
			arg1:     []string{"a", "b"},
			arg2:     []string{"a", "b"},
			expected: []string{"a", "b"},
		},
		{
			name:     "when arg1 is empty",
			arg1:     []string{},
			arg2:     []string{"a", "b"},
			expected: []string{},
		},
		{
			name:     "when arg2 is empty",
			arg1:     []string{"a", "b"},
			arg2:     []string{},
			expected: []string{},
		},
		{
			name:     "when arg1/arg2 are the different set",
			arg1:     []string{"a", "b"},
			arg2:     []string{"c", "d"},
			expected: []string{},
		},
	}

	for _, c := range cases {
		actual := filter.Intersect(c.arg1, c.arg2)
		if !assert.Equal(t, c.expected, actual) {
			t.Errorf("case: %s is failed, expected: %+v, actual: %+v\n", c.name, c.expected, actual)
		}
	}
}

func TestDifference(t *testing.T) {
	cases := []struct {
		name     string
		arg1     []string
		arg2     []string
		expected []string
	}{
		{
			name:     "when arg1/arg2 includes the mutual item",
			arg1:     []string{"a", "b"},
			arg2:     []string{"b", "c"},
			expected: []string{"a"},
		},
		{
			name:     "when arg1/arg2 are the same set",
			arg1:     []string{"a", "b"},
			arg2:     []string{"a", "b"},
			expected: []string{},
		},
		{
			name:     "when arg1 is empty",
			arg1:     []string{},
			arg2:     []string{"a", "b"},
			expected: []string{},
		},
		{
			name:     "when arg2 is empty",
			arg1:     []string{"a", "b"},
			arg2:     []string{},
			expected: []string{"a", "b"},
		},
		{
			name:     "when arg1/arg2 are the different set",
			arg1:     []string{"a", "b"},
			arg2:     []string{"c", "d"},
			expected: []string{"a", "b"},
		},
	}

	for _, c := range cases {
		actual := filter.Difference(c.arg1, c.arg2)
		if !assert.Equal(t, c.expected, actual) {
			t.Errorf("case: %s is failed, expected: %+v, actual: %+v\n", c.name, c.expected, actual)
		}
	}
}
