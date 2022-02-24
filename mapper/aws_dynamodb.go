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
)

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
func (addtm AwsDynamoDBTagsMapper) GetTagsMapping(ctx context.Context) (map[string]Tags, error) {
	if cv, found := addtm.cache.Get(awsDynamoDBCache); found {
		mapping := cv.(map[string]Tags)
		return mapping, nil
	}

	mapping := make(map[string]Tags)

	initMarker := ""
	marker := &initMarker

	for marker != nil {
		input := dynamodb.ListTablesInput{}
		output, err := addtm.client.ListTables(ctx, &input)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		for i := 0; i < len(output.TableNames); i++ {
			name := output.TableNames[i]

			tableInput := dynamodb.DescribeTableInput{TableName: &name}
			tableOutput, err := addtm.client.DescribeTable(ctx, &tableInput)
			if err != nil {
				return nil, fmt.Errorf("%w", err)
			}

			tagsInput := dynamodb.ListTagsOfResourceInput{ResourceArn: tableOutput.Table.TableArn}
			tagsOutput, err := addtm.client.ListTagsOfResource(ctx, &tagsInput)
			if err != nil {
				return nil, fmt.Errorf("%w", err)
			}

			tags := make(Tags, len(tagsOutput.Tags))
			for j, tag := range tagsOutput.Tags {
				tags[j] = fmt.Sprintf("%s:%s", strings.ToLower(*tag.Key), strings.ToLower(*tag.Value))
			}
			mapping[name] = tags
		}

		// When output.LastEvaluatedTableName is nil, it means all table names have already fetched.
		// cf. https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/dynamodb#ListTablesOutput
		marker = output.LastEvaluatedTableName
	}

	addtm.cache.Set(awsClbCache, mapping, cache.DefaultExpiration)
	return mapping, nil
}
