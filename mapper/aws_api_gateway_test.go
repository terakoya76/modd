package mapper_test

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	"github.com/aws/smithy-go/middleware"
	goCache "github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"

	"github.com/terakoya76/modd/mapper"
)

// dummyAwsAPIGatewayClient implements AwsAPIGatewayClient interface for faking AWS API.
type dummyAwsAPIGatewayClient struct{}

// GetRestApis implements AwsAPIGatewayClient for dummyAwsAPIGatewayClient.
func (c *dummyAwsAPIGatewayClient) GetRestApis(
	ctx context.Context,
	params *apigateway.GetRestApisInput,
	optFns ...func(*apigateway.Options),
) (*apigateway.GetRestApisOutput, error) {
	var output apigateway.GetRestApisOutput

	if params.Position != nil {
		output = apigateway.GetRestApisOutput{
			Items: []types.RestApi{
				{
					Name: aws.String("api10"),
					Tags: map[string]string{
						"key10": "val10",
					},
				},
				{
					Name: aws.String("api20"),
					Tags: map[string]string{
						"key20": "val20",
					},
				},
			},
			Position:       nil,
			ResultMetadata: middleware.Metadata{},
		}
	} else {
		output = apigateway.GetRestApisOutput{
			Items: []types.RestApi{
				{
					Name: aws.String("api1"),
					Tags: map[string]string{
						"key1": "val1",
					},
				},
				{
					Name: aws.String("api2"),
					Tags: map[string]string{
						"key2": "val2",
					},
				},
			},
			Position:       aws.String("next token"),
			ResultMetadata: middleware.Metadata{},
		}
	}

	return &output, nil
}

func Test_AwsAPIGateway_GetTagsMapping(t *testing.T) {
	cache := goCache.New(60*time.Minute, 10*time.Minute)

	cases := []struct {
		name     string
		expected map[string]mapper.Tags
		err      error
	}{
		{
			name: "fake test",
			expected: map[string]mapper.Tags{
				"api1":  []string{"key1:val1"},
				"api2":  []string{"key2:val2"},
				"api10": []string{"key10:val10"},
				"api20": []string{"key20:val20"},
			},
			err: nil,
		},
	}

	for _, c := range cases {
		client := dummyAwsAPIGatewayClient{}
		m := mapper.BuildAwsAPIGatewayTagsMapper(cache, &client)
		actual, err := m.GetTagsMapping(context.TODO())
		if !assert.Equal(t, c.err, err) {
			t.Errorf("case: %s is failed, expected: %+v, actual: %+v\n", c.name, c.err, err)
		}
		if !assert.Equal(t, c.expected, actual) {
			t.Errorf("case: %s is failed, expected: %+v, actual: %+v\n", c.name, c.expected, actual)
		}
	}
}
