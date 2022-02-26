package mapper

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/patrickmn/go-cache"

	"github.com/terakoya76/modd/datadog"
)

const awsDynamoDBCacheKey string = string(datadog.AwsDynamoDB)

// AwsDynamoDBTagsMapper implements TagsMapper for AWS DynamoDB.
type AwsDynamoDBTagsMapper struct {
	cache  *cache.Cache
	client *dynamodb.Client
}

// GetAwsDynamoDBClient returns AWS DynamoDB client.
func GetAwsDynamoDBClient(ctx context.Context) (*dynamodb.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRetryer(func() aws.Retryer {
		return retry.AddWithMaxAttempts(retry.NewStandard(), 10)
	}))
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return dynamodb.NewFromConfig(cfg), nil
}

// GetTagsMapping returns the latest tags mapping.
func (tm AwsDynamoDBTagsMapper) GetTagsMapping(ctx context.Context) (map[string]Tags, error) {
	if cv, found := tm.cache.Get(awsDynamoDBCacheKey); found {
		mapping := cv.(map[string]Tags)
		return mapping, nil
	}

	mapping := make(map[string]Tags)

	// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/dynamodb#ListTablesInput
	initMarker := ""
	marker := &initMarker

	for marker != nil {
		// ExclusiveStartTableName could not be empty string
		var input dynamodb.ListTablesInput
		if *marker == "" {
			input = dynamodb.ListTablesInput{}
		} else {
			input = dynamodb.ListTablesInput{ExclusiveStartTableName: marker}
		}

		output, err := tm.client.ListTables(ctx, &input)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		for i := 0; i < len(output.TableNames); i++ {
			name := output.TableNames[i]

			tableInput := dynamodb.DescribeTableInput{TableName: &name}
			tableOutput, err := tm.client.DescribeTable(ctx, &tableInput)
			if err != nil {
				return nil, fmt.Errorf("%w", err)
			}

			tags := make(Tags, 0)

			// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/dynamodb#ListTagsOfResourceInput
			initTagMarker := ""
			tagMarker := &initTagMarker

			for tagMarker != nil {
				// NextToken could not be empty string
				var tagsInput dynamodb.ListTagsOfResourceInput
				if *tagMarker == "" {
					tagsInput = dynamodb.ListTagsOfResourceInput{ResourceArn: tableOutput.Table.TableArn}
				} else {
					tagsInput = dynamodb.ListTagsOfResourceInput{ResourceArn: tableOutput.Table.TableArn, NextToken: tagMarker}
				}

				tagsOutput, err := tm.client.ListTagsOfResource(ctx, &tagsInput)
				if err != nil {
					return nil, fmt.Errorf("%w", err)
				}

				for _, tag := range tagsOutput.Tags {
					tags = append(tags, fmt.Sprintf("%s:%s", strings.ToLower(*tag.Key), strings.ToLower(*tag.Value)))
				}

				tagMarker = tagsOutput.NextToken
			}

			mapping[name] = tags
		}

		// When output.LastEvaluatedTableName is nil, it means all table names have already fetched.
		// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/dynamodb#ListTablesOutput
		marker = output.LastEvaluatedTableName
	}

	tm.cache.Set(awsDynamoDBCacheKey, mapping, cache.DefaultExpiration)
	return mapping, nil
}
