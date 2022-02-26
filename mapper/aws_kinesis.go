package mapper

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/patrickmn/go-cache"

	"github.com/terakoya76/modd/datadog"
)

const awsKinesisCacheKey string = string(datadog.AwsKinesis)

// AwsKinesisTagsMapper implements TagsMapper for AWS Kinesis.
type AwsKinesisTagsMapper struct {
	cache  *cache.Cache
	client *kinesis.Client
}

// GetAwsKinesisClient returns AWS Kinesis client.
func GetAwsKinesisClient(ctx context.Context) (*kinesis.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRetryer(func() aws.Retryer {
		return retry.AddWithMaxAttempts(retry.NewStandard(), 10)
	}))
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return kinesis.NewFromConfig(cfg), nil
}

// GetTagsMapping returns the latest tags mapping.
func (tm AwsKinesisTagsMapper) GetTagsMapping(ctx context.Context) (map[string]Tags, error) {
	if cv, found := tm.cache.Get(awsKinesisCacheKey); found {
		mapping := cv.(map[string]Tags)
		return mapping, nil
	}

	mapping := make(map[string]Tags)

	// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/kinesis#ListStreamsInput
	initMarker := ""
	lastReturnedStreamName := &initMarker

	// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/kinesis#ListStreamsOutput
	hasMoreStream := true
	for hasMoreStream {
		// ExclusiveStartStreamName could not be empty string
		var input kinesis.ListStreamsInput
		if *lastReturnedStreamName == "" {
			input = kinesis.ListStreamsInput{}
		} else {
			input = kinesis.ListStreamsInput{ExclusiveStartStreamName: lastReturnedStreamName}
		}

		output, err := tm.client.ListStreams(ctx, &input)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		returnedStreamNamesCount := len(output.StreamNames)
		if returnedStreamNamesCount > 0 {
			lastReturnedStreamName = &output.StreamNames[returnedStreamNamesCount-1]
		}

		for i := 0; i < returnedStreamNamesCount; i++ {
			name := output.StreamNames[i]

			tags := make(Tags, 0)

			// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/kinesis#ListTagsForStreamInput
			initTagMarker := ""
			lastReturnedTagKey := &initTagMarker

			// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/kinesis#ListTagsForStreamOutput
			hasMoreTag := true
			for hasMoreTag {
				// ExclusiveStartTagKey could not be empty string
				var tagsInput kinesis.ListTagsForStreamInput
				if *lastReturnedTagKey == "" {
					tagsInput = kinesis.ListTagsForStreamInput{StreamName: &name}
				} else {
					tagsInput = kinesis.ListTagsForStreamInput{StreamName: &name, ExclusiveStartTagKey: lastReturnedTagKey}
				}

				tagsOutput, err := tm.client.ListTagsForStream(ctx, &tagsInput)
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

		hasMoreStream = *output.HasMoreStreams
	}

	tm.cache.Set(awsKinesisCacheKey, mapping, cache.DefaultExpiration)
	return mapping, nil
}
