package mapper

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/patrickmn/go-cache"

	"github.com/terakoya76/modd/datadog"
)

const awsLambdaCacheKey string = string(datadog.AwsLambda)

// AwsLambdaTagsMapper implements TagsMapper for AWS Lambda.
type AwsLambdaTagsMapper struct {
	cache  *cache.Cache
	client *lambda.Client
}

// GetAwsLambdaClient returns AWS Lambda client.
func GetAwsLambdaClient(ctx context.Context) (*lambda.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRetryer(func() aws.Retryer {
		return retry.AddWithMaxAttempts(retry.NewStandard(), 10)
	}))
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return lambda.NewFromConfig(cfg), nil
}

// GetTagsMapping returns the latest tags mapping.
func (altm AwsLambdaTagsMapper) GetTagsMapping(ctx context.Context) (map[string]Tags, error) {
	if cv, found := altm.cache.Get(awsLambdaCacheKey); found {
		mapping := cv.(map[string]Tags)
		return mapping, nil
	}

	mapping := make(map[string]Tags)

	// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/lambda#ListFunctionsInput
	initMarker := ""
	marker := &initMarker

	for marker != nil {
		// Marker could not be empty string
		var input lambda.ListFunctionsInput
		if *marker == "" {
			input = lambda.ListFunctionsInput{}
		} else {
			input = lambda.ListFunctionsInput{Marker: marker}
		}

		output, err := altm.client.ListFunctions(ctx, &input)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		for i := 0; i < len(output.Functions); i++ {
			function := output.Functions[i]
			tagsInput := lambda.ListTagsInput{Resource: function.FunctionArn}
			tagsOutput, err := altm.client.ListTags(ctx, &tagsInput)
			if err != nil {
				return nil, fmt.Errorf("%w", err)
			}

			tags := make(Tags, len(tagsOutput.Tags))
			j := 0
			for k, v := range tagsOutput.Tags {
				tags[j] = fmt.Sprintf("%s:%s", strings.ToLower(k), strings.ToLower(v))
				j++
			}

			mapping[*function.FunctionName] = tags
		}

		marker = output.NextMarker
	}

	altm.cache.Set(awsLambdaCacheKey, mapping, cache.DefaultExpiration)
	return mapping, nil
}
