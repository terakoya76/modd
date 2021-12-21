package mapper

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/patrickmn/go-cache"
)

// AwsRdsTagsMapper implements TagsMapper for AWS RDS.
type AwsRdsTagsMapper struct {
	cache  *cache.Cache
	client *rds.Client
}

// GetAwsRdsClient returns AWS RDS client.
func GetAwsRdsClient(ctx context.Context) (*rds.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRetryer(func() aws.Retryer {
		return retry.AddWithMaxAttempts(retry.NewStandard(), 10)
	}))
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return rds.NewFromConfig(cfg), nil
}

// GetTagsMapping returns the latest tags mapping.
func (artm AwsRdsTagsMapper) GetTagsMapping(ctx context.Context) (map[string]Tags, error) {
	if cv, found := artm.cache.Get(awsRdsCache); found {
		mapping := cv.(map[string]Tags)
		return mapping, nil
	}

	mapping := make(map[string]Tags)

	initMarker := ""
	marker := &initMarker

	for marker != nil {
		input := rds.DescribeDBInstancesInput{Marker: marker}
		output, err := artm.client.DescribeDBInstances(ctx, &input)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		for i := 0; i < len(output.DBInstances); i++ {
			db := output.DBInstances[i]
			tags := make(Tags, len(db.TagList))
			for j, tag := range db.TagList {
				tags[j] = fmt.Sprintf("%s:%s", strings.ToLower(*tag.Key), strings.ToLower(*tag.Value))
			}
			mapping[*db.DBInstanceIdentifier] = tags
		}

		marker = output.Marker
	}

	artm.cache.Set(awsRdsCache, mapping, cache.DefaultExpiration)
	return mapping, nil
}
