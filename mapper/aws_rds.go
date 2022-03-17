package mapper

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	goCache "github.com/patrickmn/go-cache"

	"github.com/terakoya76/modd/datadog"
)

const awsRdsCacheKey string = string(datadog.AwsRds)

// AwsRdsClient is abstract interface of *rds.Client.
type AwsRdsClient interface {
	DescribeDBInstances(
		ctx context.Context,
		params *rds.DescribeDBInstancesInput,
		optFns ...func(*rds.Options),
	) (*rds.DescribeDBInstancesOutput, error)
}

// AwsRdsTagsMapper implements TagsMapper for AWS RDS.
type AwsRdsTagsMapper struct {
	cache  *goCache.Cache
	client AwsRdsClient
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

// BuildAwsRdsTagsMapper builds AwsRdsTagsMapper from args.
func BuildAwsRdsTagsMapper(cache *goCache.Cache, client AwsRdsClient) AwsRdsTagsMapper {
	return AwsRdsTagsMapper{
		cache:  cache,
		client: client,
	}
}

// GetTagsMapping returns the latest tags mapping.
func (tm AwsRdsTagsMapper) GetTagsMapping(ctx context.Context) (map[string]Tags, error) {
	if cv, found := tm.cache.Get(awsRdsCacheKey); found {
		mapping := cv.(map[string]Tags)
		return mapping, nil
	}

	mapping := make(map[string]Tags)

	// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/rds#DescribeDBInstancesInput
	marker := aws.String("")

	for marker != nil {
		// Marker could not be empty string
		var input rds.DescribeDBInstancesInput
		if *marker == "" {
			input = rds.DescribeDBInstancesInput{}
		} else {
			input = rds.DescribeDBInstancesInput{Marker: marker}
		}

		output, err := tm.client.DescribeDBInstances(ctx, &input)
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

	tm.cache.Set(awsRdsCacheKey, mapping, goCache.DefaultExpiration)
	return mapping, nil
}
