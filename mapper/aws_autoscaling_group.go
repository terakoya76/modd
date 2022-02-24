package mapper

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/patrickmn/go-cache"

	"github.com/terakoya76/modd/datadog"
)

const awsAutoScalingGroupCacheKey string = string(datadog.AwsAutoScalingGroup)

// AwsAutoScalingGroupTagsMapper implements TagsMapper for AWS AutoScalingGroup.
type AwsAutoScalingGroupTagsMapper struct {
	cache  *cache.Cache
	client *autoscaling.Client
}

// GetAwsAutoScalingGroupClient returns AWS AutoScalingGroup client.
func GetAwsAutoScalingGroupClient(ctx context.Context) (*autoscaling.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRetryer(func() aws.Retryer {
		return retry.AddWithMaxAttempts(retry.NewStandard(), 10)
	}))
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return autoscaling.NewFromConfig(cfg), nil
}

// GetTagsMapping returns the latest tags mapping.
func (aasgtm AwsAutoScalingGroupTagsMapper) GetTagsMapping(ctx context.Context) (map[string]Tags, error) {
	if cv, found := aasgtm.cache.Get(awsAutoScalingGroupCacheKey); found {
		mapping := cv.(map[string]Tags)
		return mapping, nil
	}

	mapping := make(map[string]Tags)

	// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/autoscaling#DescribeAutoScalingGroupsInput
	initToken := ""
	token := &initToken

	for token != nil {
		// NextToken could not be empty string
		var input autoscaling.DescribeAutoScalingGroupsInput
		if *token == "" {
			input = autoscaling.DescribeAutoScalingGroupsInput{}
		} else {
			input = autoscaling.DescribeAutoScalingGroupsInput{NextToken: token}
		}

		output, err := aasgtm.client.DescribeAutoScalingGroups(ctx, &input)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		for i := 0; i < len(output.AutoScalingGroups); i++ {
			asg := output.AutoScalingGroups[i]
			tags := make(Tags, len(asg.Tags))
			for j, tag := range asg.Tags {
				tags[j] = fmt.Sprintf("%s:%s", strings.ToLower(*tag.Key), strings.ToLower(*tag.Value))
			}
			mapping[*asg.AutoScalingGroupName] = tags
		}

		token = output.NextToken
	}

	aasgtm.cache.Set(awsAutoScalingGroupCacheKey, mapping, cache.DefaultExpiration)
	return mapping, nil
}
