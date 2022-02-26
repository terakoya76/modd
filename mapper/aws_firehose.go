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
func (aftm AwsFirehoseTagsMapper) GetTagsMapping(ctx context.Context) (map[string]Tags, error) {
	if cv, found := aftm.cache.Get(awsFirehoseCacheKey); found {
		mapping := cv.(map[string]Tags)
		return mapping, nil
	}

	mapping := make(map[string]Tags)

	// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/firehose#ListDeliveryStreamsOutput
	hasMoreStream := true
	for hasMoreStream {
		input := firehose.ListDeliveryStreamsInput{}

		output, err := aftm.client.ListDeliveryStreams(ctx, &input)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		for i := 0; i < len(output.DeliveryStreamNames); i++ {
			name := output.DeliveryStreamNames[i]

			tagsInput := firehose.ListTagsForDeliveryStreamInput{DeliveryStreamName: &name}
			tagsOutput, err := aftm.client.ListTagsForDeliveryStream(ctx, &tagsInput)
			if err != nil {
				return nil, fmt.Errorf("%w", err)
			}

			tags := make(Tags, len(tagsOutput.Tags))
			for j, tag := range tagsOutput.Tags {
				tags[j] = fmt.Sprintf("%s:%s", strings.ToLower(*tag.Key), strings.ToLower(*tag.Value))
			}
			mapping[name] = tags
		}

		hasMoreStream = *output.HasMoreDeliveryStreams
	}

	aftm.cache.Set(awsFirehoseCacheKey, mapping, cache.DefaultExpiration)
	return mapping, nil
}
