package evaluator_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/terakoya76/modd/evaluator"
)

func Test_GetIdentifiersFromMaaping(t *testing.T) {
	cases := []struct {
		name     string
		mapping  map[string][]string
		expected []string
	}{
		{
			name: "when scope is wildcard without tags",
			mapping: map[string][]string{
				"foo": {"a", "b"},
				"bar": {"c", "d"},
			},
			expected: []string{"foo", "bar"},
		},
	}

	for _, c := range cases {
		actual := evaluator.GetIdentifiersFromMaaping(c.mapping)
		if !assert.ElementsMatch(t, c.expected, actual) {
			t.Errorf("case: %s is failed, expected: %+v, actual: %+v\n", c.name, c.expected, actual)
		}
	}
}
