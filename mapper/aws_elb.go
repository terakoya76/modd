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
)

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
func (aetm AwsElbTagsMapper) GetTagsMapping(ctx context.Context) (map[string]Tags, error) {
	if cv, found := aetm.cache.Get(awsElbCache); found {
		mapping := cv.(map[string]Tags)
		return mapping, nil
	}

	mapping := make(map[string]Tags)

	initMarker := ""
	marker := &initMarker

	for marker != nil {
		input := elasticloadbalancingv2.DescribeLoadBalancersInput{}
		output, err := aetm.client.DescribeLoadBalancers(ctx, &input)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		for i := 0; i < len(output.LoadBalancers); i++ {
			lb := output.LoadBalancers[i]
			tagsInput := elasticloadbalancingv2.DescribeTagsInput{ResourceArns: []string{*lb.LoadBalancerArn}}
			tagsOutput, err := aetm.client.DescribeTags(ctx, &tagsInput)
			if err != nil {
				return nil, fmt.Errorf("%w", err)
			}

			for j := 0; j < len(tagsOutput.TagDescriptions); j++ {
				tags := make(Tags, len(tagsOutput.TagDescriptions[j].Tags))
				for k, tag := range tagsOutput.TagDescriptions[j].Tags {
					tags[k] = fmt.Sprintf("%s:%s", strings.ToLower(*tag.Key), strings.ToLower(*tag.Value))
				}
				mapping[*lb.LoadBalancerName] = tags
			}
		}

		marker = output.NextMarker
	}

	aetm.cache.Set(awsElbCache, mapping, cache.DefaultExpiration)
	return mapping, nil
}
