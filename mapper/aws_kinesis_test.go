package mapper_test

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/aws/aws-sdk-go-v2/service/kinesis/types"
	"github.com/aws/smithy-go/middleware"
	goCache "github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"

	"github.com/terakoya76/modd/mapper"
)

// dummyAwsKinesisClient implements AwsKinesisClient interface for faking AWS API.
type dummyAwsKinesisClient struct{}

// ListStreams implements AwsKinesisClient for dummyAwsKinesisClient.
func (c *dummyAwsKinesisClient) ListStreams(
	_ context.Context,
	params *kinesis.ListStreamsInput,
	_ ...func(*kinesis.Options),
) (*kinesis.ListStreamsOutput, error) {
	var output kinesis.ListStreamsOutput

	if params.ExclusiveStartStreamName != nil {
		output = kinesis.ListStreamsOutput{
			HasMoreStreams: aws.Bool(false),
			StreamNames:    []string{"stream10", "stream20"},
			ResultMetadata: middleware.Metadata{},
		}
	} else {
		output = kinesis.ListStreamsOutput{
			HasMoreStreams: aws.Bool(true),
			StreamNames:    []string{"stream1", "stream2"},
			ResultMetadata: middleware.Metadata{},
		}
	}

	return &output, nil
}

// ListTagsForStream implements AwsKinesisClient for dummyAwsKinesisClient.
func (c *dummyAwsKinesisClient) ListTagsForStream(
	_ context.Context,
	params *kinesis.ListTagsForStreamInput,
	_ ...func(*kinesis.Options),
) (*kinesis.ListTagsForStreamOutput, error) {
	var output kinesis.ListTagsForStreamOutput

	if params.ExclusiveStartTagKey != nil {
		output = kinesis.ListTagsForStreamOutput{
			HasMoreTags: aws.Bool(false),
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
		output = kinesis.ListTagsForStreamOutput{
			HasMoreTags: aws.Bool(true),
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

func Test_AwsKinesis_GetTagsMapping(t *testing.T) {
	cache := goCache.New(60*time.Minute, 10*time.Minute)

	cases := []struct {
		name     string
		expected map[string]mapper.Tags
		err      error
	}{
		{
			name: "fake test",
			expected: map[string]mapper.Tags{
				"stream1":  []string{"key1:val1", "key2:val2", "key10:val10", "key20:val20"},
				"stream2":  []string{"key1:val1", "key2:val2", "key10:val10", "key20:val20"},
				"stream10": []string{"key1:val1", "key2:val2", "key10:val10", "key20:val20"},
				"stream20": []string{"key1:val1", "key2:val2", "key10:val10", "key20:val20"},
			},
			err: nil,
		},
	}

	for _, c := range cases {
		client := dummyAwsKinesisClient{}
		m := mapper.BuildAwsKinesisTagsMapper(cache, &client)
		actual, err := m.GetTagsMapping(context.TODO())
		if !assert.Equal(t, c.err, err) {
			t.Errorf("case: %s is failed, expected: %+v, actual: %+v\n", c.name, c.err, err)
		}
		if !assert.Equal(t, c.expected, actual) {
			t.Errorf("case: %s is failed, expected: %+v, actual: %+v\n", c.name, c.expected, actual)
		}
	}
}
