package mapper

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	"github.com/patrickmn/go-cache"

	"github.com/terakoya76/modd/datadog"
)

const awsClbCacheKey string = string(datadog.AwsClb)

// AwsClbTagsMapper implements TagsMapper for AWS CLB.
type AwsClbTagsMapper struct {
	cache  *cache.Cache
	client *elasticloadbalancing.Client
}

// GetAwsClbClient returns AWS CLB client.
func GetAwsClbClient(ctx context.Context) (*elasticloadbalancing.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRetryer(func() aws.Retryer {
		return retry.AddWithMaxAttempts(retry.NewStandard(), 10)
	}))
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return elasticloadbalancing.NewFromConfig(cfg), nil
}

// GetTagsMapping returns the latest tags mapping.
func (actm AwsClbTagsMapper) GetTagsMapping(ctx context.Context) (map[string]Tags, error) {
	if cv, found := actm.cache.Get(awsClbCacheKey); found {
		mapping := cv.(map[string]Tags)
		return mapping, nil
	}

	mapping := make(map[string]Tags)

	// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing#DescribeLoadBalancersInput
	initMarker := ""
	marker := &initMarker

	// Although there is no description, we can fetch load balancers upto 20 in a single call.
	// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing#DescribeLoadBalancersInput
	maxItemsPerReq := 20

	for marker != nil {
		// Marker could not be empty string
		var input elasticloadbalancing.DescribeLoadBalancersInput
		if *marker == "" {
			input = elasticloadbalancing.DescribeLoadBalancersInput{}
		} else {
			input = elasticloadbalancing.DescribeLoadBalancersInput{Marker: marker}
		}

		output, err := actm.client.DescribeLoadBalancers(ctx, &input)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		iter := len(output.LoadBalancerDescriptions)/maxItemsPerReq + 1
		for i := 0; i < iter; i++ {
			names := []string{}
			for j := 0; j < maxItemsPerReq; j++ {
				idx := iter*i + j
				if j >= len(output.LoadBalancerDescriptions) {
					continue
				}

				lb := output.LoadBalancerDescriptions[idx]
				names = append(names, *lb.LoadBalancerName)
			}

			tagsInput := elasticloadbalancing.DescribeTagsInput{LoadBalancerNames: names}
			tagsOutput, err := actm.client.DescribeTags(ctx, &tagsInput)
			if err != nil {
				return nil, fmt.Errorf("%w", err)
			}

			for j := 0; j < len(tagsOutput.TagDescriptions); j++ {
				tags := make(Tags, len(tagsOutput.TagDescriptions[j].Tags))
				for k, tag := range tagsOutput.TagDescriptions[j].Tags {
					tags[k] = fmt.Sprintf("%s:%s", strings.ToLower(*tag.Key), strings.ToLower(*tag.Value))
				}

				idx := iter*i + j
				lb := output.LoadBalancerDescriptions[idx]
				mapping[*lb.LoadBalancerName] = tags
			}
		}

		marker = output.NextMarker
	}

	actm.cache.Set(awsClbCacheKey, mapping, cache.DefaultExpiration)
	return mapping, nil
}
