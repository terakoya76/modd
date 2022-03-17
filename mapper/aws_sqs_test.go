package mapper_test

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/smithy-go/middleware"
	goCache "github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"

	"github.com/terakoya76/modd/mapper"
)

// dummyAwsSqsClient implements AwsSqsClient interface for faking AWS API.
type dummyAwsSqsClient struct{}

// ListQueues implements AwsSqsClient for dummyAwsSqsClient.
func (c *dummyAwsSqsClient) ListQueues(
	ctx context.Context,
	params *sqs.ListQueuesInput,
	optFns ...func(*sqs.Options),
) (*sqs.ListQueuesOutput, error) {
	var output sqs.ListQueuesOutput

	if params.NextToken != nil {
		output = sqs.ListQueuesOutput{
			QueueUrls: []string{
				"https://sqs/queue10",
				"https://sqs/queue20",
			},
			NextToken:      nil,
			ResultMetadata: middleware.Metadata{},
		}
	} else {
		output = sqs.ListQueuesOutput{
			QueueUrls: []string{
				"https://sqs/queue1",
				"https://sqs/queue2",
			},
			NextToken:      aws.String("next token"),
			ResultMetadata: middleware.Metadata{},
		}
	}

	return &output, nil
}

// ListQueueTags implements AwsSqsClient for dummyAwsSqsClient.
func (c *dummyAwsSqsClient) ListQueueTags(
	ctx context.Context,
	params *sqs.ListQueueTagsInput,
	optFns ...func(*sqs.Options),
) (*sqs.ListQueueTagsOutput, error) {
	output := sqs.ListQueueTagsOutput{
		Tags: map[string]string{
			"key1": "val1",
			"key2": "val2",
		},
		ResultMetadata: middleware.Metadata{},
	}

	return &output, nil
}

func Test_AwsSqs_GetTagsMapping(t *testing.T) {
	cache := goCache.New(60*time.Minute, 10*time.Minute)

	cases := []struct {
		name     string
		expected map[string]mapper.Tags
		err      error
	}{
		{
			name: "fake test",
			expected: map[string]mapper.Tags{
				"queue1":  []string{"key1:val1", "key2:val2"},
				"queue2":  []string{"key1:val1", "key2:val2"},
				"queue10": []string{"key1:val1", "key2:val2"},
				"queue20": []string{"key1:val1", "key2:val2"},
			},
			err: nil,
		},
	}

	for _, c := range cases {
		client := dummyAwsSqsClient{}
		m := mapper.BuildAwsSqsTagsMapper(cache, &client)
		actual, err := m.GetTagsMapping(context.TODO())
		if !assert.Equal(t, c.err, err) {
			t.Errorf("case: %s is failed, expected: %+v, actual: %+v\n", c.name, c.err, err)
		}

		for k, v := range c.expected {
			if !assert.ElementsMatch(t, v, actual[k]) {
				t.Errorf("case: %s is failed with the key %s, expected: %+v, actual: %+v\n", c.name, k, v, actual[k])
			}
		}
	}
}
