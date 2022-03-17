package mapper_test

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/aws/smithy-go/middleware"
	goCache "github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"

	"github.com/terakoya76/modd/mapper"
)

// dummyAwsElbClient implements AwsElbClient interface for faking AWS API.
type dummyAwsElbClient struct{}

// DescribeLoadBalancers implements AwsElbClient for dummyAwsElbClient.
func (c *dummyAwsElbClient) DescribeLoadBalancers(
	ctx context.Context,
	params *elasticloadbalancingv2.DescribeLoadBalancersInput,
	optFns ...func(*elasticloadbalancingv2.Options),
) (*elasticloadbalancingv2.DescribeLoadBalancersOutput, error) {
	var output elasticloadbalancingv2.DescribeLoadBalancersOutput

	if params.Marker != nil {
		output = elasticloadbalancingv2.DescribeLoadBalancersOutput{
			LoadBalancers: []types.LoadBalancer{
				{
					LoadBalancerArn:  aws.String("lb arn 10"),
					LoadBalancerName: aws.String("lb10"),
				},
				{
					LoadBalancerArn:  aws.String("lb arn 20"),
					LoadBalancerName: aws.String("lb20"),
				},
			},
			NextMarker:     nil,
			ResultMetadata: middleware.Metadata{},
		}
	} else {
		output = elasticloadbalancingv2.DescribeLoadBalancersOutput{
			LoadBalancers: []types.LoadBalancer{
				{
					LoadBalancerArn:  aws.String("lb arn 1"),
					LoadBalancerName: aws.String("lb1"),
				},
				{
					LoadBalancerArn:  aws.String("lb arn 2"),
					LoadBalancerName: aws.String("lb2"),
				},
			},
			NextMarker:     aws.String("next marker"),
			ResultMetadata: middleware.Metadata{},
		}
	}

	return &output, nil
}

// DescribeTags implements AwsElbClient for dummyAwsElbClient.
func (c *dummyAwsElbClient) DescribeTags(
	ctx context.Context,
	params *elasticloadbalancingv2.DescribeTagsInput,
	optFns ...func(*elasticloadbalancingv2.Options),
) (*elasticloadbalancingv2.DescribeTagsOutput, error) {
	tags := make([]types.TagDescription, len(params.ResourceArns))

	for i, arn := range params.ResourceArns {
		tags[i] = types.TagDescription{
			ResourceArn: aws.String(arn),
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
		}
	}

	output := elasticloadbalancingv2.DescribeTagsOutput{
		TagDescriptions: tags,
		ResultMetadata:  middleware.Metadata{},
	}

	return &output, nil
}

func Test_AwsElb_GetTagsMapping(t *testing.T) {
	cache := goCache.New(60*time.Minute, 10*time.Minute)

	cases := []struct {
		name     string
		expected map[string]mapper.Tags
		err      error
	}{
		{
			name: "fake test",
			expected: map[string]mapper.Tags{
				"lb1":  []string{"key1:val1", "key2:val2"},
				"lb2":  []string{"key1:val1", "key2:val2"},
				"lb10": []string{"key1:val1", "key2:val2"},
				"lb20": []string{"key1:val1", "key2:val2"},
			},
			err: nil,
		},
	}

	for _, c := range cases {
		client := dummyAwsElbClient{}
		m := mapper.BuildAwsElbTagsMapper(cache, &client)
		actual, err := m.GetTagsMapping(context.TODO())
		if !assert.Equal(t, c.err, err) {
			t.Errorf("case: %s is failed, expected: %+v, actual: %+v\n", c.name, c.err, err)
		}

		for k, v := range c.expected {
			if !assert.ElementsMatch(t, v, actual[k]) {
				t.Errorf("case: %s is failed with the key %s, expected: %+v, actual: %+v\n", c.name, k, v, actual[k])
			}
		}
	}
}
