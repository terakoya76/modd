package datadog_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/terakoya76/modd/datadog"
)

func Test_MakeUniq(t *testing.T) {
	cases := []struct {
		name     string
		arr      [][]string
		expected [][]string
	}{
		{
			name: "when scope is wildcard without tags",
			arr: [][]string{
				{"a"},
				{"a"},
			},
			expected: [][]string{
				{"a"},
			},
		},
		{
			name: "when scope is wildcard without tags",
			arr: [][]string{
				{"a"},
				{"b"},
			},
			expected: [][]string{
				{"a"},
				{"b"},
			},
		},
		{
			name: "when scope is wildcard without tags",
			arr: [][]string{
				{"a"},
				{"b"},
				{"a"},
				{"b"},
			},
			expected: [][]string{
				{"a"},
				{"b"},
			},
		},
	}

	for _, c := range cases {
		actual := datadog.MakeUniq(c.arr)
		if !assert.ElementsMatch(t, c.expected, actual) {
			t.Errorf("case: %s is failed, expected: %+v, actual: %+v\n", c.name, c.expected, actual)
		}
	}
}
