package mapper_test

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
	"github.com/aws/smithy-go/middleware"
	goCache "github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"

	"github.com/terakoya76/modd/mapper"
)

// dummyAwsAutoScalingGroupClient implements AwsAutoScalingGroupClient interface for faking AWS API.
type dummyAwsAutoScalingGroupClient struct{}

// DescribeAutoScalingGroups implements AwsAutoScalingGroupClient for dummyAwsAutoScalingGroupClient.
func (c *dummyAwsAutoScalingGroupClient) DescribeAutoScalingGroups(
	_ context.Context,
	params *autoscaling.DescribeAutoScalingGroupsInput,
	_ ...func(*autoscaling.Options),
) (*autoscaling.DescribeAutoScalingGroupsOutput, error) {
	var output autoscaling.DescribeAutoScalingGroupsOutput

	if params.NextToken != nil {
		output = autoscaling.DescribeAutoScalingGroupsOutput{
			AutoScalingGroups: []types.AutoScalingGroup{
				{
					AutoScalingGroupName: aws.String("gateway10"),
					Tags: []types.TagDescription{
						{
							Key:   aws.String("key10"),
							Value: aws.String("val10"),
						},
					},
				},
				{
					AutoScalingGroupName: aws.String("gateway20"),
					Tags: []types.TagDescription{
						{
							Key:   aws.String("key20"),
							Value: aws.String("val20"),
						},
					},
				},
			},
			NextToken:      nil,
			ResultMetadata: middleware.Metadata{},
		}
	} else {
		output = autoscaling.DescribeAutoScalingGroupsOutput{
			AutoScalingGroups: []types.AutoScalingGroup{
				{
					AutoScalingGroupName: aws.String("gateway1"),
					Tags: []types.TagDescription{
						{
							Key:   aws.String("key1"),
							Value: aws.String("val1"),
						},
					},
				},
				{
					AutoScalingGroupName: aws.String("gateway2"),
					Tags: []types.TagDescription{
						{
							Key:   aws.String("key2"),
							Value: aws.String("val2"),
						},
					},
				},
			},
			NextToken:      aws.String("next token"),
			ResultMetadata: middleware.Metadata{},
		}
	}

	return &output, nil
}

func Test_AwsAutoScalingGroup_GetTagsMapping(t *testing.T) {
	cache := goCache.New(60*time.Minute, 10*time.Minute)

	cases := []struct {
		name     string
		expected map[string]mapper.Tags
		err      error
	}{
		{
			name: "fake test",
			expected: map[string]mapper.Tags{
				"gateway1":  []string{"key1:val1"},
				"gateway2":  []string{"key2:val2"},
				"gateway10": []string{"key10:val10"},
				"gateway20": []string{"key20:val20"},
			},
			err: nil,
		},
	}

	for _, c := range cases {
		client := dummyAwsAutoScalingGroupClient{}
		m := mapper.BuildAwsAutoScalingGroupTagsMapper(cache, &client)
		actual, err := m.GetTagsMapping(context.TODO())
		if !assert.Equal(t, c.err, err) {
			t.Errorf("case: %s is failed, expected: %+v, actual: %+v\n", c.name, c.err, err)
		}
		if !assert.Equal(t, c.expected, actual) {
			t.Errorf("case: %s is failed, expected: %+v, actual: %+v\n", c.name, c.expected, actual)
		}
	}
}
