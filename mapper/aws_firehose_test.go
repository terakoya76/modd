package mapper_test

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/firehose"
	"github.com/aws/aws-sdk-go-v2/service/firehose/types"
	"github.com/aws/smithy-go/middleware"
	goCache "github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"

	"github.com/terakoya76/modd/mapper"
)

// dummyAwsFirehoseClient implements AwsFirehoseClient interface for faking AWS API.
type dummyAwsFirehoseClient struct{}

// ListStreams implements AwsFirehoseClient for dummyAwsFirehoseClient.
func (c *dummyAwsFirehoseClient) ListDeliveryStreams(
	ctx context.Context,
	params *firehose.ListDeliveryStreamsInput,
	optFns ...func(*firehose.Options),
) (*firehose.ListDeliveryStreamsOutput, error) {
	var output firehose.ListDeliveryStreamsOutput

	if params.ExclusiveStartDeliveryStreamName != nil {
		output = firehose.ListDeliveryStreamsOutput{
			HasMoreDeliveryStreams: aws.Bool(false),
			DeliveryStreamNames:    []string{"stream10", "stream20"},
			ResultMetadata:         middleware.Metadata{},
		}
	} else {
		output = firehose.ListDeliveryStreamsOutput{
			HasMoreDeliveryStreams: aws.Bool(true),
			DeliveryStreamNames:    []string{"stream1", "stream2"},
			ResultMetadata:         middleware.Metadata{},
		}
	}

	return &output, nil
}

// ListTagsForStream implements AwsFirehoseClient for dummyAwsFirehoseClient.
func (c *dummyAwsFirehoseClient) ListTagsForDeliveryStream(
	ctx context.Context,
	params *firehose.ListTagsForDeliveryStreamInput,
	optFns ...func(*firehose.Options),
) (*firehose.ListTagsForDeliveryStreamOutput, error) {
	var output firehose.ListTagsForDeliveryStreamOutput

	if params.ExclusiveStartTagKey != nil {
		output = firehose.ListTagsForDeliveryStreamOutput{
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
		output = firehose.ListTagsForDeliveryStreamOutput{
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

func Test_AwsFirehose_GetTagsMapping(t *testing.T) {
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
		client := dummyAwsFirehoseClient{}
		m := mapper.BuildAwsFirehoseTagsMapper(cache, &client)
		actual, err := m.GetTagsMapping(context.TODO())
		if !assert.Equal(t, c.err, err) {
			t.Errorf("case: %s is failed, expected: %+v, actual: %+v\n", c.name, c.err, err)
		}
		if !assert.Equal(t, c.expected, actual) {
			t.Errorf("case: %s is failed, expected: %+v, actual: %+v\n", c.name, c.expected, actual)
		}
	}
}
