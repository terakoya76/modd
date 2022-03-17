package mapper

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	goCache "github.com/patrickmn/go-cache"

	"github.com/terakoya76/modd/datadog"
)

const awsKinesisCacheKey string = string(datadog.AwsKinesis)

// AwsKinesisClient is abstract interface of *kinesis.Client.
type AwsKinesisClient interface {
	ListStreams(
		ctx context.Context,
		params *kinesis.ListStreamsInput,
		optFns ...func(*kinesis.Options),
	) (*kinesis.ListStreamsOutput, error)
	ListTagsForStream(
		ctx context.Context,
		params *kinesis.ListTagsForStreamInput,
		optFns ...func(*kinesis.Options),
	) (*kinesis.ListTagsForStreamOutput, error)
}

// AwsKinesisTagsMapper implements TagsMapper for AWS Kinesis.
type AwsKinesisTagsMapper struct {
	cache  *goCache.Cache
	client AwsKinesisClient
}

// BuildAwsKinesisTagsMapper builds AwsKinesisTagsMapper from args.
func BuildAwsKinesisTagsMapper(cache *goCache.Cache, client AwsKinesisClient) AwsKinesisTagsMapper {
	return AwsKinesisTagsMapper{
		cache:  cache,
		client: client,
	}
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
	lastReturnedStreamName := aws.String("")

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
			lastReturnedTagKey := aws.String("")

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

	tm.cache.Set(awsKinesisCacheKey, mapping, goCache.DefaultExpiration)
	return mapping, nil
}
