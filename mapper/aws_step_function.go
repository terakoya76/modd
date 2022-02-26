package mapper

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
	"github.com/patrickmn/go-cache"

	"github.com/terakoya76/modd/datadog"
)

const awsStepFunctionCacheKey string = string(datadog.AwsStepFunction)

// AwsStepFunctionTagsMapper implements TagsMapper for AWS StepFunction.
type AwsStepFunctionTagsMapper struct {
	cache  *cache.Cache
	client *sfn.Client
}

// GetAwsStepFunctionClient returns AWS StepFunction client.
func GetAwsStepFunctionClient(ctx context.Context) (*sfn.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRetryer(func() aws.Retryer {
		return retry.AddWithMaxAttempts(retry.NewStandard(), 10)
	}))
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return sfn.NewFromConfig(cfg), nil
}

// GetTagsMapping returns the latest tags mapping.
func (tm AwsStepFunctionTagsMapper) GetTagsMapping(ctx context.Context) (map[string]Tags, error) {
	if cv, found := tm.cache.Get(awsStepFunctionCacheKey); found {
		mapping := cv.(map[string]Tags)
		return mapping, nil
	}

	mapping := make(map[string]Tags)

	// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/sfn#ListStateMachinesInput
	initToken := ""
	token := &initToken

	// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/sfn#ListStateMachinesInput
	var maxResults int32 = 1000

	for token != nil {
		// NextToken could not be empty string
		var input sfn.ListStateMachinesInput
		if *token == "" {
			input = sfn.ListStateMachinesInput{MaxResults: maxResults}
		} else {
			input = sfn.ListStateMachinesInput{MaxResults: maxResults, NextToken: token}
		}

		output, err := tm.client.ListStateMachines(ctx, &input)

		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		for i := 0; i < len(output.StateMachines); i++ {
			sm := output.StateMachines[i]
			tagsInput := sfn.ListTagsForResourceInput{ResourceArn: sm.StateMachineArn}
			tagsOutput, err := tm.client.ListTagsForResource(ctx, &tagsInput)
			if err != nil {
				return nil, fmt.Errorf("%w", err)
			}

			tags := make(Tags, len(tagsOutput.Tags))
			for j, tag := range tagsOutput.Tags {
				tags[j] = fmt.Sprintf("%s:%s", strings.ToLower(*tag.Key), strings.ToLower(*tag.Value))
			}
			mapping[*sm.Name] = tags
		}

		token = output.NextToken
	}

	tm.cache.Set(awsStepFunctionCacheKey, mapping, cache.DefaultExpiration)
	return mapping, nil
}
