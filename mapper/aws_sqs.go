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
)

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
func (astm AwsSqsTagsMapper) GetTagsMapping(ctx context.Context) (map[string]Tags, error) {
	if cv, found := astm.cache.Get(awsSqsCache); found {
		mapping := cv.(map[string]Tags)
		return mapping, nil
	}

	mapping := make(map[string]Tags)

	initToken := ""
	token := &initToken

	var maxResult int32 = 1000

	for token != nil {
		var input sqs.ListQueuesInput
		if *token == "" {
			input = sqs.ListQueuesInput{MaxResults: &maxResult}
		} else {
			input = sqs.ListQueuesInput{MaxResults: &maxResult, NextToken: token}
		}

		output, err := astm.client.ListQueues(ctx, &input)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		for i := 0; i < len(output.QueueUrls); i++ {
			queueURL := output.QueueUrls[i]
			tagsInput := sqs.ListQueueTagsInput{QueueUrl: &queueURL}
			tagsOutput, err := astm.client.ListQueueTags(ctx, &tagsInput)
			if err != nil {
				return nil, fmt.Errorf("%w", err)
			}

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

	astm.cache.Set(awsSqsCache, mapping, cache.DefaultExpiration)
	return mapping, nil
}
