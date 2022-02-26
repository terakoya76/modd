package mapper

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/patrickmn/go-cache"

	"github.com/terakoya76/modd/datadog"
)

const awsAPIGatewayCacheKey string = string(datadog.AwsAPIGateway)

// AwsAPIGatewayTagsMapper implements TagsMapper for AWS API Gateway.
type AwsAPIGatewayTagsMapper struct {
	cache  *cache.Cache
	client *apigateway.Client
}

// GetAwsAPIGatewayClient returns AWS API Gateway client.
func GetAwsAPIGatewayClient(ctx context.Context) (*apigateway.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRetryer(func() aws.Retryer {
		return retry.AddWithMaxAttempts(retry.NewStandard(), 10)
	}))
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return apigateway.NewFromConfig(cfg), nil
}

// GetTagsMapping returns the latest tags mapping.
func (tm AwsAPIGatewayTagsMapper) GetTagsMapping(ctx context.Context) (map[string]Tags, error) {
	if cv, found := tm.cache.Get(awsAPIGatewayCacheKey); found {
		mapping := cv.(map[string]Tags)
		return mapping, nil
	}

	mapping := make(map[string]Tags)

	// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/apigateway#GetRestApisInput
	initPos := ""
	pos := &initPos

	// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/apigateway#GetRestApisInput
	var limit int32 = 500

	for pos != nil {
		// Position could not be empty string
		var input apigateway.GetRestApisInput
		if *pos == "" {
			input = apigateway.GetRestApisInput{Limit: &limit}
		} else {
			input = apigateway.GetRestApisInput{Limit: &limit, Position: pos}
		}

		output, err := tm.client.GetRestApis(ctx, &input)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		for i := 0; i < len(output.Items); i++ {
			api := output.Items[i]

			// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/apigateway@v1.14.0/types#RestApi
			tags := make(Tags, len(api.Tags))
			j := 0
			for k, v := range api.Tags {
				tags[j] = fmt.Sprintf("%s:%s", strings.ToLower(k), strings.ToLower(v))
				j++
			}

			mapping[*api.Name] = tags
		}

		pos = output.Position
	}

	tm.cache.Set(awsAPIGatewayCacheKey, mapping, cache.DefaultExpiration)
	return mapping, nil
}
