package mapper

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	goCache "github.com/patrickmn/go-cache"

	"github.com/terakoya76/modd/datadog"
)

const awsElbCacheKey string = string(datadog.AwsElb)

// AwsElbClient is abstract interface of *elasticloadbalancingv2.Client.
type AwsElbClient interface {
	DescribeLoadBalancers(
		ctx context.Context,
		params *elasticloadbalancingv2.DescribeLoadBalancersInput,
		optFns ...func(*elasticloadbalancingv2.Options),
	) (*elasticloadbalancingv2.DescribeLoadBalancersOutput, error)
	DescribeTags(
		ctx context.Context,
		params *elasticloadbalancingv2.DescribeTagsInput,
		optFns ...func(*elasticloadbalancingv2.Options),
	) (*elasticloadbalancingv2.DescribeTagsOutput, error)
}

// AwsElbTagsMapper implements TagsMapper for AWS ALB/NLB.
type AwsElbTagsMapper struct {
	cache  *goCache.Cache
	client AwsElbClient
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

// BuildAwsElbTagsMapper builds AwsElbTagsMapper from args.
func BuildAwsElbTagsMapper(cache *goCache.Cache, client AwsElbClient) AwsElbTagsMapper {
	return AwsElbTagsMapper{
		cache:  cache,
		client: client,
	}
}

// GetTagsMapping returns the latest tags mapping.
func (tm AwsElbTagsMapper) GetTagsMapping(ctx context.Context) (map[string]Tags, error) {
	if cv, found := tm.cache.Get(awsElbCacheKey); found {
		mapping := cv.(map[string]Tags)
		return mapping, nil
	}

	mapping := make(map[string]Tags)

	// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2#DescribeLoadBalancersInput
	marker := aws.String("")

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
				if j >= len(output.LoadBalancers) {
					continue
				}

				idx := iter*i + j
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

	tm.cache.Set(awsElbCacheKey, mapping, goCache.DefaultExpiration)
	return mapping, nil
}
