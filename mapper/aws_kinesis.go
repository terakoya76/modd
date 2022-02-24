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
func (aktm AwsKinesisTagsMapper) GetTagsMapping(ctx context.Context) (map[string]Tags, error) {
	if cv, found := aktm.cache.Get(awsKinesisCacheKey); found {
		mapping := cv.(map[string]Tags)
		return mapping, nil
	}

	mapping := make(map[string]Tags)

	// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/kinesis#ListStreamsInput
	hasMoreStream := true
	for hasMoreStream {
		input := kinesis.ListStreamsInput{}

		output, err := aktm.client.ListStreams(ctx, &input)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		for i := 0; i < len(output.StreamNames); i++ {
			name := output.StreamNames[i]
			tagsInput := kinesis.ListTagsForStreamInput{StreamName: &name}
			tagsOutput, err := aktm.client.ListTagsForStream(ctx, &tagsInput)
			if err != nil {
				return nil, fmt.Errorf("%w", err)
			}

			tags := make(Tags, len(tagsOutput.Tags))
			for j, tag := range tagsOutput.Tags {
				tags[j] = fmt.Sprintf("%s:%s", strings.ToLower(*tag.Key), strings.ToLower(*tag.Value))
			}
			mapping[name] = tags
		}

		hasMoreStream = *output.HasMoreStreams
	}

	aktm.cache.Set(awsKinesisCacheKey, mapping, cache.DefaultExpiration)
	return mapping, nil
}
