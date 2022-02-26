package mapper

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/patrickmn/go-cache"

	"github.com/terakoya76/modd/datadog"
)

const awsElbCacheKey string = string(datadog.AwsElb)

// AwsElbTagsMapper implements TagsMapper for AWS ALB/NLB.
type AwsElbTagsMapper struct {
	cache  *cache.Cache
	client *elasticloadbalancingv2.Client
}

// GetAwsElbClient returns AWS ALB/NLB client.
func GetAwsElbClient(ctx context.Context) (*elasticloadbalancingv2.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRetryer(func() aws.Retryer {
		return retry.AddWithMaxAttempts(retry.NewStandard(), 10)
	}))
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return elasticloadbalancingv2.NewFromConfig(cfg), nil
}

// GetTagsMapping returns the latest tags mapping.
func (tm AwsElbTagsMapper) GetTagsMapping(ctx context.Context) (map[string]Tags, error) {
	if cv, found := tm.cache.Get(awsElbCacheKey); found {
		mapping := cv.(map[string]Tags)
		return mapping, nil
	}

	mapping := make(map[string]Tags)

	// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2#DescribeLoadBalancersInput
	initMarker := ""
	marker := &initMarker

	// We can fetch load balancers upto 20 in a single call.
	// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2#DescribeLoadBalancersInput
	maxItemsPerReq := 20

	for marker != nil {
		// Marker could not be empty string
		var input elasticloadbalancingv2.DescribeLoadBalancersInput
		if *marker == "" {
			input = elasticloadbalancingv2.DescribeLoadBalancersInput{}
		} else {
			input = elasticloadbalancingv2.DescribeLoadBalancersInput{Marker: marker}
		}

		output, err := tm.client.DescribeLoadBalancers(ctx, &input)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		iter := len(output.LoadBalancers)/maxItemsPerReq + 1
		for i := 0; i < iter; i++ {
			arns := []string{}
			for j := 0; j < maxItemsPerReq; j++ {
				idx := iter*i + j
				if j >= len(output.LoadBalancers) {
					continue
				}

				lb := output.LoadBalancers[idx]
				arns = append(arns, *lb.LoadBalancerArn)
			}

			tagsInput := elasticloadbalancingv2.DescribeTagsInput{ResourceArns: arns}
			tagsOutput, err := tm.client.DescribeTags(ctx, &tagsInput)
			if err != nil {
				return nil, fmt.Errorf("%w", err)
			}

			for j := 0; j < len(tagsOutput.TagDescriptions); j++ {
				tags := make(Tags, len(tagsOutput.TagDescriptions[j].Tags))
				for k, tag := range tagsOutput.TagDescriptions[j].Tags {
					tags[k] = fmt.Sprintf("%s:%s", strings.ToLower(*tag.Key), strings.ToLower(*tag.Value))
				}

				idx := iter*i + j
				lb := output.LoadBalancers[idx]
				mapping[*lb.LoadBalancerName] = tags
			}
		}

		marker = output.NextMarker
	}

	tm.cache.Set(awsElbCacheKey, mapping, cache.DefaultExpiration)
	return mapping, nil
}
