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
)

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
	if cv, found := actm.cache.Get(awsClbCache); found {
		mapping := cv.(map[string]Tags)
		return mapping, nil
	}

	mapping := make(map[string]Tags)

	initMarker := ""
	marker := &initMarker

	for marker != nil {
		input := elasticloadbalancing.DescribeLoadBalancersInput{}
		output, err := actm.client.DescribeLoadBalancers(ctx, &input)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		for i := 0; i < len(output.LoadBalancerDescriptions); i++ {
			lb := output.LoadBalancerDescriptions[i]
			tagsInput := elasticloadbalancing.DescribeTagsInput{LoadBalancerNames: []string{*lb.LoadBalancerName}}
			tagsOutput, err := actm.client.DescribeTags(ctx, &tagsInput)
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

	actm.cache.Set(awsClbCache, mapping, cache.DefaultExpiration)
	return mapping, nil
}
