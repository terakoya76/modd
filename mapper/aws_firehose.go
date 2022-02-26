package mapper

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/firehose"
	"github.com/patrickmn/go-cache"

	"github.com/terakoya76/modd/datadog"
)

const awsFirehoseCacheKey string = string(datadog.AwsFirehose)

// AwsFirehoseTagsMapper implements TagsMapper for AWS Firehose.
type AwsFirehoseTagsMapper struct {
	cache  *cache.Cache
	client *firehose.Client
}

// GetAwsFirehoseClient returns AWS Firehose client.
func GetAwsFirehoseClient(ctx context.Context) (*firehose.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRetryer(func() aws.Retryer {
		return retry.AddWithMaxAttempts(retry.NewStandard(), 10)
	}))
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return firehose.NewFromConfig(cfg), nil
}

// GetTagsMapping returns the latest tags mapping.
func (tm AwsFirehoseTagsMapper) GetTagsMapping(ctx context.Context) (map[string]Tags, error) {
	if cv, found := tm.cache.Get(awsFirehoseCacheKey); found {
		mapping := cv.(map[string]Tags)
		return mapping, nil
	}

	mapping := make(map[string]Tags)

	// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/firehose#ListDeliveryStreamsInput
	initMarker := ""
	lastReturnedStreamName := &initMarker

	// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/firehose#ListDeliveryStreamsOutput
	hasMoreStream := true
	for hasMoreStream {
		// ExclusiveStartDeliveryStreamName could not be empty string
		var input firehose.ListDeliveryStreamsInput
		if *lastReturnedStreamName == "" {
			input = firehose.ListDeliveryStreamsInput{}
		} else {
			input = firehose.ListDeliveryStreamsInput{ExclusiveStartDeliveryStreamName: lastReturnedStreamName}
		}

		output, err := tm.client.ListDeliveryStreams(ctx, &input)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		returnedStreamNamesCount := len(output.DeliveryStreamNames)
		if returnedStreamNamesCount > 0 {
			lastReturnedStreamName = &output.DeliveryStreamNames[returnedStreamNamesCount-1]
		}

		for i := 0; i < returnedStreamNamesCount; i++ {
			name := output.DeliveryStreamNames[i]

			tags := make(Tags, 0)

			// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/firehose#ListTagsForDeliveryStreamInput
			initTagMarker := ""
			lastReturnedTagKey := &initTagMarker

			// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/firehose#ListTagsForDeliveryStreamOutput
			hasMoreTag := true
			for hasMoreTag {
				// ExclusiveStartTagKey could not be empty string
				var tagsInput firehose.ListTagsForDeliveryStreamInput
				if *lastReturnedTagKey == "" {
					tagsInput = firehose.ListTagsForDeliveryStreamInput{DeliveryStreamName: &name}
				} else {
					tagsInput = firehose.ListTagsForDeliveryStreamInput{DeliveryStreamName: &name, ExclusiveStartTagKey: lastReturnedTagKey}
				}

				tagsOutput, err := tm.client.ListTagsForDeliveryStream(ctx, &tagsInput)
				if err != nil {
					return nil, fmt.Errorf("%w", err)
				}

				for _, tag := range tagsOutput.Tags {
					tags = append(tags, fmt.Sprintf("%s:%s", strings.ToLower(*tag.Key), strings.ToLower(*tag.Value)))
				}

				returnedTagsCount := len(tagsOutput.Tags)
				if returnedTagsCount > 0 {
					lastReturnedTagKey = tagsOutput.Tags[returnedTagsCount-1].Key
				}

				hasMoreTag = *tagsOutput.HasMoreTags
			}

			mapping[name] = tags
		}

		hasMoreStream = *output.HasMoreDeliveryStreams
	}

	tm.cache.Set(awsFirehoseCacheKey, mapping, cache.DefaultExpiration)
	return mapping, nil
}
