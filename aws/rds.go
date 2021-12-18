package aws

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/rds"
)

// Tags represents AWS tags
type Tags = []string

// RdsTagsMapping represents a mapping of AWS RDS identifier and its tags
type RdsTagsMapping = map[string]Tags

// GetRdsClient returns AWS RDS client
func GetRdsClient(ctx context.Context) (*rds.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return rds.NewFromConfig(cfg), nil
}

// GetRdsTagsMapping returns the latest RdsTagsMapping
func GetRdsTagsMapping(ctx context.Context, rdsClient *rds.Client) (RdsTagsMapping, error) {
	mapping := make(RdsTagsMapping)

	initMarker := ""
	var marker *string = &initMarker

	for marker != nil {
		input := rds.DescribeDBInstancesInput{Marker: marker}
		output, err := rdsClient.DescribeDBInstances(ctx, &input)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		for _, db := range output.DBInstances {
			tags := make(Tags, len(db.TagList), len(db.TagList))
			for i, tag := range db.TagList {
				tags[i] = fmt.Sprintf("%s:%s", strings.ToLower(*tag.Key), strings.ToLower(*tag.Value))
			}
			mapping[*db.DBInstanceIdentifier] = tags
		}

		marker = output.Marker
	}

	return mapping, nil
}

// GetRdsIdentifiers returns a list of AWS RDS identifiers
func GetRdsIdentifiers(mapping RdsTagsMapping) []string {
	keys := make([]string, len(mapping))
	i := 0
	for k := range mapping {
		keys[i] = k
		i++
	}

	return keys
}
