package mapper_test

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/smithy-go/middleware"
	goCache "github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"

	"github.com/terakoya76/modd/mapper"
)

// dummyAwsDynamoDBClient implements AwsDynamoDBClient interface for faking AWS API.
type dummyAwsDynamoDBClient struct{}

// DescribeTable implements AwsDynamoDBClient for dummyAwsDynamoDBClient.
func (c *dummyAwsDynamoDBClient) DescribeTable(
	_ context.Context,
	_ *dynamodb.DescribeTableInput,
	_ ...func(*dynamodb.Options),
) (*dynamodb.DescribeTableOutput, error) {
	output := dynamodb.DescribeTableOutput{
		Table: &types.TableDescription{
			TableArn: aws.String("arn"),
		},
		ResultMetadata: middleware.Metadata{},
	}

	return &output, nil
}

// ListTables implements AwsDynamoDBClient for dummyAwsDynamoDBClient.
func (c *dummyAwsDynamoDBClient) ListTables(
	_ context.Context,
	params *dynamodb.ListTablesInput,
	_ ...func(*dynamodb.Options),
) (*dynamodb.ListTablesOutput, error) {
	var output dynamodb.ListTablesOutput

	if params.ExclusiveStartTableName != nil {
		output = dynamodb.ListTablesOutput{
			LastEvaluatedTableName: nil,
			TableNames:             []string{"table10", "table20"},
			ResultMetadata:         middleware.Metadata{},
		}
	} else {
		output = dynamodb.ListTablesOutput{
			LastEvaluatedTableName: aws.String("last table"),
			TableNames:             []string{"table1", "table2"},
			ResultMetadata:         middleware.Metadata{},
		}
	}

	return &output, nil
}

// ListTagsOfResource implements AwsDynamoDBClient for dummyAwsDynamoDBClient.
func (c *dummyAwsDynamoDBClient) ListTagsOfResource(
	_ context.Context,
	params *dynamodb.ListTagsOfResourceInput,
	_ ...func(*dynamodb.Options),
) (*dynamodb.ListTagsOfResourceOutput, error) {
	var output dynamodb.ListTagsOfResourceOutput

	if params.NextToken != nil {
		output = dynamodb.ListTagsOfResourceOutput{
			NextToken: nil,
			Tags: []types.Tag{
				{
					Key:   aws.String("key10"),
					Value: aws.String("val10"),
				},
				{
					Key:   aws.String("key20"),
					Value: aws.String("val20"),
				},
			},
			ResultMetadata: middleware.Metadata{},
		}
	} else {
		output = dynamodb.ListTagsOfResourceOutput{
			NextToken: aws.String("next token"),
			Tags: []types.Tag{
				{
					Key:   aws.String("key1"),
					Value: aws.String("val1"),
				},
				{
					Key:   aws.String("key2"),
					Value: aws.String("val2"),
				},
			},
			ResultMetadata: middleware.Metadata{},
		}
	}

	return &output, nil
}

func Test_AwsDynamoDB_GetTagsMapping(t *testing.T) {
	cache := goCache.New(60*time.Minute, 10*time.Minute)

	cases := []struct {
		name     string
		expected map[string]mapper.Tags
		err      error
	}{
		{
			name: "fake test",
			expected: map[string]mapper.Tags{
				"table1":  []string{"key1:val1", "key2:val2", "key10:val10", "key20:val20"},
				"table2":  []string{"key1:val1", "key2:val2", "key10:val10", "key20:val20"},
				"table10": []string{"key1:val1", "key2:val2", "key10:val10", "key20:val20"},
				"table20": []string{"key1:val1", "key2:val2", "key10:val10", "key20:val20"},
			},
			err: nil,
		},
	}

	for _, c := range cases {
		client := dummyAwsDynamoDBClient{}
		m := mapper.BuildAwsDynamoDBTagsMapper(cache, &client)
		actual, err := m.GetTagsMapping(context.TODO())
		if !assert.Equal(t, c.err, err) {
			t.Errorf("case: %s is failed, expected: %+v, actual: %+v\n", c.name, c.err, err)
		}
		if !assert.Equal(t, c.expected, actual) {
			t.Errorf("case: %s is failed, expected: %+v, actual: %+v\n", c.name, c.expected, actual)
		}
	}
}
