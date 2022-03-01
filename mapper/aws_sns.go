package mapper

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/patrickmn/go-cache"

	"github.com/terakoya76/modd/datadog"
)

const awsSnsCacheKey string = string(datadog.AwsSns)

// AwsSnsTagsMapper implements TagsMapper for AWS SNS.
type AwsSnsTagsMapper struct {
	cache  *cache.Cache
	client *sns.Client
}

// GetAwsSnsClient returns AWS SNS client.
func GetAwsSnsClient(ctx context.Context) (*sns.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRetryer(func() aws.Retryer {
		return retry.AddWithMaxAttempts(retry.NewStandard(), 10)
	}))
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return sns.NewFromConfig(cfg), nil
}

// GetTagsMapping returns the latest tags mapping.
func (tm AwsSnsTagsMapper) GetTagsMapping(ctx context.Context) (map[string]Tags, error) {
	if cv, found := tm.cache.Get(awsSnsCacheKey); found {
		mapping := cv.(map[string]Tags)
		return mapping, nil
	}

	mapping := make(map[string]Tags)

	// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/sns#ListTopicsInput
	initToken := ""
	token := &initToken

	for token != nil {
		// NextToken could not be empty string
		var input sns.ListTopicsInput
		if *token == "" {
			input = sns.ListTopicsInput{}
		} else {
			input = sns.ListTopicsInput{NextToken: token}
		}

		output, err := tm.client.ListTopics(ctx, &input)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		for i := 0; i < len(output.Topics); i++ {
			topic := output.Topics[i]
			tagsInput := sns.ListTagsForResourceInput{ResourceArn: topic.TopicArn}
			tagsOutput, err := tm.client.ListTagsForResource(ctx, &tagsInput)
			if err != nil {
				return nil, fmt.Errorf("%w", err)
			}

			tags := make(Tags, len(tagsOutput.Tags))
			for j, tag := range tagsOutput.Tags {
				tags[j] = fmt.Sprintf("%s:%s", strings.ToLower(*tag.Key), strings.ToLower(*tag.Value))
			}

			arn := *topic.TopicArn
			name := arn[strings.LastIndex(arn, ":")+1:]
			mapping[name] = tags
		}

		token = output.NextToken
	}

	tm.cache.Set(awsSnsCacheKey, mapping, cache.DefaultExpiration)
	return mapping, nil
}
