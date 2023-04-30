package mapper_test

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"
	"github.com/aws/smithy-go/middleware"
	goCache "github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"

	"github.com/terakoya76/modd/mapper"
)

// dummyAwsClbClient implements AwsClbClient interface for faking AWS API.
type dummyAwsClbClient struct{}

// DescribeLoadBalancers implements AwsClbClient for dummyAwsClbClient.
func (c *dummyAwsClbClient) DescribeLoadBalancers(
	_ context.Context,
	params *elasticloadbalancing.DescribeLoadBalancersInput,
	_ ...func(*elasticloadbalancing.Options),
) (*elasticloadbalancing.DescribeLoadBalancersOutput, error) {
	var output elasticloadbalancing.DescribeLoadBalancersOutput

	if params.Marker != nil {
		output = elasticloadbalancing.DescribeLoadBalancersOutput{
			LoadBalancerDescriptions: []types.LoadBalancerDescription{
				{
					LoadBalancerName: aws.String("lb10"),
				},
				{
					LoadBalancerName: aws.String("lb20"),
				},
			},
			NextMarker:     nil,
			ResultMetadata: middleware.Metadata{},
		}
	} else {
		output = elasticloadbalancing.DescribeLoadBalancersOutput{
			LoadBalancerDescriptions: []types.LoadBalancerDescription{
				{
					LoadBalancerName: aws.String("lb1"),
				},
				{
					LoadBalancerName: aws.String("lb2"),
				},
			},
			NextMarker:     aws.String("next marker"),
			ResultMetadata: middleware.Metadata{},
		}
	}

	return &output, nil
}

// DescribeTags implements AwsClbClient for dummyAwsClbClient.
func (c *dummyAwsClbClient) DescribeTags(
	_ context.Context,
	params *elasticloadbalancing.DescribeTagsInput,
	_ ...func(*elasticloadbalancing.Options),
) (*elasticloadbalancing.DescribeTagsOutput, error) {
	tags := make([]types.TagDescription, len(params.LoadBalancerNames))

	for i, name := range params.LoadBalancerNames {
		tags[i] = types.TagDescription{
			LoadBalancerName: aws.String(name),
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

	output := elasticloadbalancing.DescribeTagsOutput{
		TagDescriptions: tags,
		ResultMetadata:  middleware.Metadata{},
	}

	return &output, nil
}

func Test_AwsClb_GetTagsMapping(t *testing.T) {
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
		client := dummyAwsClbClient{}
		m := mapper.BuildAwsClbTagsMapper(cache, &client)
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
