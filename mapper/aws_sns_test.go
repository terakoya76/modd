package mapper_test

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/aws/smithy-go/middleware"
	goCache "github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"

	"github.com/terakoya76/modd/mapper"
)

// dummyAwsSnsClient implements AwsSnsClient interface for faking AWS API.
type dummyAwsSnsClient struct{}

// ListTopics implements AwsSnsClient for dummyAwsSnsClient.
func (c *dummyAwsSnsClient) ListTopics(
	ctx context.Context,
	params *sns.ListTopicsInput,
	optFns ...func(*sns.Options),
) (*sns.ListTopicsOutput, error) {
	var output sns.ListTopicsOutput

	if params.NextToken != nil {
		output = sns.ListTopicsOutput{
			Topics: []types.Topic{
				{
					TopicArn: aws.String("xxx:topic10"),
				},
				{
					TopicArn: aws.String("xxx:topic20"),
				},
			},
			NextToken:      nil,
			ResultMetadata: middleware.Metadata{},
		}
	} else {
		output = sns.ListTopicsOutput{
			Topics: []types.Topic{
				{
					TopicArn: aws.String("xxx:topic1"),
				},
				{
					TopicArn: aws.String("xxx:topic2"),
				},
			},
			NextToken:      aws.String("next token"),
			ResultMetadata: middleware.Metadata{},
		}
	}

	return &output, nil
}

// ListTagsForResource implements AwsSnsClient for dummyAwsSnsClient.
func (c *dummyAwsSnsClient) ListTagsForResource(
	ctx context.Context,
	params *sns.ListTagsForResourceInput,
	optFns ...func(*sns.Options),
) (*sns.ListTagsForResourceOutput, error) {
	output := sns.ListTagsForResourceOutput{
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

	return &output, nil
}

func Test_AwsSns_GetTagsMapping(t *testing.T) {
	cache := goCache.New(60*time.Minute, 10*time.Minute)

	cases := []struct {
		name     string
		expected map[string]mapper.Tags
		err      error
	}{
		{
			name: "fake test",
			expected: map[string]mapper.Tags{
				"topic1":  []string{"key1:val1", "key2:val2"},
				"topic2":  []string{"key1:val1", "key2:val2"},
				"topic10": []string{"key1:val1", "key2:val2"},
				"topic20": []string{"key1:val1", "key2:val2"},
			},
			err: nil,
		},
	}

	for _, c := range cases {
		client := dummyAwsSnsClient{}
		m := mapper.BuildAwsSnsTagsMapper(cache, &client)
		actual, err := m.GetTagsMapping(context.TODO())
		if !assert.Equal(t, c.err, err) {
			t.Errorf("case: %s is failed, expected: %+v, actual: %+v\n", c.name, c.err, err)
		}
		if !assert.Equal(t, c.expected, actual) {
			t.Errorf("case: %s is failed, expected: %+v, actual: %+v\n", c.name, c.expected, actual)
		}
	}
}
