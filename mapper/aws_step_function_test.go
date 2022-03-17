package mapper_test

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
	"github.com/aws/aws-sdk-go-v2/service/sfn/types"
	"github.com/aws/smithy-go/middleware"
	goCache "github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"

	"github.com/terakoya76/modd/mapper"
)

// dummyAwsStepFunctionClient implements AwsStepFunctionClient interface for faking AWS API.
type dummyAwsStepFunctionClient struct{}

// ListStateMachines implements AwsStepFunctionClient for dummyAwsStepFunctionClient.
func (c *dummyAwsStepFunctionClient) ListStateMachines(
	ctx context.Context,
	params *sfn.ListStateMachinesInput,
	optFns ...func(*sfn.Options),
) (*sfn.ListStateMachinesOutput, error) {
	var output sfn.ListStateMachinesOutput

	if params.NextToken != nil {
		output = sfn.ListStateMachinesOutput{
			StateMachines: []types.StateMachineListItem{
				{
					Name:            aws.String("sfn10"),
					StateMachineArn: aws.String("sfn arn 10"),
				},
				{
					Name:            aws.String("sfn20"),
					StateMachineArn: aws.String("sfn arn 20"),
				},
			},
			NextToken:      nil,
			ResultMetadata: middleware.Metadata{},
		}
	} else {
		output = sfn.ListStateMachinesOutput{
			StateMachines: []types.StateMachineListItem{
				{
					Name:            aws.String("sfn1"),
					StateMachineArn: aws.String("sfn arn 1"),
				},
				{
					Name:            aws.String("sfn2"),
					StateMachineArn: aws.String("sfn arn 2"),
				},
			},
			NextToken:      aws.String("next token"),
			ResultMetadata: middleware.Metadata{},
		}
	}

	return &output, nil
}

// ListTagsForResource implements AwsStepFunctionClient for dummyAwsStepFunctionClient.
func (c *dummyAwsStepFunctionClient) ListTagsForResource(
	ctx context.Context,
	params *sfn.ListTagsForResourceInput,
	optFns ...func(*sfn.Options),
) (*sfn.ListTagsForResourceOutput, error) {
	output := sfn.ListTagsForResourceOutput{
		Tags: []types.Tag{
			{
				Key:   aws.String("key1"),
				Value: aws.String("val1"),
			},
			{
				Key:   aws.String("key2"),
				Value: aws.String("val2"),
			},
		},
		ResultMetadata: middleware.Metadata{},
	}

	return &output, nil
}

func Test_AwsStepFunction_GetTagsMapping(t *testing.T) {
	cache := goCache.New(60*time.Minute, 10*time.Minute)

	cases := []struct {
		name     string
		expected map[string]mapper.Tags
		err      error
	}{
		{
			name: "fake test",
			expected: map[string]mapper.Tags{
				"sfn1":  []string{"key1:val1", "key2:val2"},
				"sfn2":  []string{"key1:val1", "key2:val2"},
				"sfn10": []string{"key1:val1", "key2:val2"},
				"sfn20": []string{"key1:val1", "key2:val2"},
			},
			err: nil,
		},
	}

	for _, c := range cases {
		client := dummyAwsStepFunctionClient{}
		m := mapper.BuildAwsStepFunctionTagsMapper(cache, &client)
		actual, err := m.GetTagsMapping(context.TODO())
		if !assert.Equal(t, c.err, err) {
			t.Errorf("case: %s is failed, expected: %+v, actual: %+v\n", c.name, c.err, err)
		}
		if !assert.Equal(t, c.expected, actual) {
			t.Errorf("case: %s is failed, expected: %+v, actual: %+v\n", c.name, c.expected, actual)
		}
	}
}
