package mapper

import (
	"context"
	"fmt"

	"github.com/terakoya76/modd/datadog"
)

// Tags represents resource tags.
type Tags = []string

// TagsMapper is an interface to fetch resources and map their ids and tags.
type TagsMapper interface {
	GetTagsMapping(ctx context.Context) (map[string]Tags, error)
}

// BuildTagsMapper build the proper TagsMapper implementation.
func BuildTagsMapper(it datadog.IntegrationTarget) (TagsMapper, error) {
	switch it {
	case datadog.AwsRds:
		return AwsRdsTagsMapper{}, nil
	case datadog.AwsElasticache:
		return AwsElasticacheTagsMapper{}, nil
	default:
		return nil, fmt.Errorf("unsupported IntegrationTarget")
	}
}
