package mapper

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/patrickmn/go-cache"

	"github.com/terakoya76/modd/datadog"
)

const awsSqsCacheKey string = string(datadog.AwsSqs)

// AwsSqsTagsMapper implements TagsMapper for AWS SQS.
type AwsSqsTagsMapper struct {
	cache  *cache.Cache
	client *sqs.Client
}

// GetAwsSqsClient returns AWS SQS client.
func GetAwsSqsClient(ctx context.Context) (*sqs.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRetryer(func() aws.Retryer {
		return retry.AddWithMaxAttempts(retry.NewStandard(), 10)
	}))
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return sqs.NewFromConfig(cfg), nil
}

// GetTagsMapping returns the latest tags mapping.
func (tm AwsSqsTagsMapper) GetTagsMapping(ctx context.Context) (map[string]Tags, error) {
	if cv, found := tm.cache.Get(awsSqsCacheKey); found {
		mapping := cv.(map[string]Tags)
		return mapping, nil
	}

	mapping := make(map[string]Tags)

	// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/sqs#ListQueuesInput
	initToken := ""
	token := &initToken

	// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/sqs#ListQueuesInput
	var maxResults int32 = 1000

	for token != nil {
		// NextToken could not be empty string
		var input sqs.ListQueuesInput
		if *token == "" {
			input = sqs.ListQueuesInput{MaxResults: &maxResults}
		} else {
			input = sqs.ListQueuesInput{MaxResults: &maxResults, NextToken: token}
		}

		output, err := tm.client.ListQueues(ctx, &input)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		for i := 0; i < len(output.QueueUrls); i++ {
			queueURL := output.QueueUrls[i]
			tagsInput := sqs.ListQueueTagsInput{QueueUrl: &queueURL}
			tagsOutput, err := tm.client.ListQueueTags(ctx, &tagsInput)
			if err != nil {
				return nil, fmt.Errorf("%w", err)
			}

			// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/sqs#ListQueueTagsOutput
			tags := make(Tags, len(tagsOutput.Tags))
			j := 0
			for k, v := range tagsOutput.Tags {
				tags[j] = fmt.Sprintf("%s:%s", strings.ToLower(k), strings.ToLower(v))
				j++
			}

			queueName := queueURL[strings.LastIndex(queueURL, "/")+1:]
			mapping[queueName] = tags
		}

		token = output.NextToken
	}

	tm.cache.Set(awsSqsCacheKey, mapping, cache.DefaultExpiration)
	return mapping, nil
}
