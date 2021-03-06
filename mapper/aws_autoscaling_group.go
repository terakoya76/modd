package mapper

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	goCache "github.com/patrickmn/go-cache"

	"github.com/terakoya76/modd/datadog"
)

const awsAutoScalingGroupCacheKey string = string(datadog.AwsAutoScalingGroup)

// AwsAutoScalingGroupClient is abstract interface of *autoscaling.Client.
type AwsAutoScalingGroupClient interface {
	DescribeAutoScalingGroups(
		ctx context.Context,
		params *autoscaling.DescribeAutoScalingGroupsInput,
		optFns ...func(*autoscaling.Options),
	) (*autoscaling.DescribeAutoScalingGroupsOutput, error)
}

// AwsAutoScalingGroupTagsMapper implements TagsMapper for AWS AutoScalingGroup.
type AwsAutoScalingGroupTagsMapper struct {
	cache  *goCache.Cache
	client AwsAutoScalingGroupClient
}

// BuildAwsAutoScalingGroupTagsMapper builds AwsAutoScalingGroupTagsMapper from args.
func BuildAwsAutoScalingGroupTagsMapper(cache *goCache.Cache, client AwsAutoScalingGroupClient) AwsAutoScalingGroupTagsMapper {
	return AwsAutoScalingGroupTagsMapper{
		cache:  cache,
		client: client,
	}
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
func (tm AwsAutoScalingGroupTagsMapper) GetTagsMapping(ctx context.Context) (map[string]Tags, error) {
	if cv, found := tm.cache.Get(awsAutoScalingGroupCacheKey); found {
		mapping := cv.(map[string]Tags)
		return mapping, nil
	}

	mapping := make(map[string]Tags)

	// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/autoscaling#DescribeAutoScalingGroupsInput
	token := aws.String("")

	for token != nil {
		// NextToken could not be empty string
		var input autoscaling.DescribeAutoScalingGroupsInput
		if *token == "" {
			input = autoscaling.DescribeAutoScalingGroupsInput{}
		} else {
			input = autoscaling.DescribeAutoScalingGroupsInput{NextToken: token}
		}

		output, err := tm.client.DescribeAutoScalingGroups(ctx, &input)
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

	tm.cache.Set(awsAutoScalingGroupCacheKey, mapping, goCache.DefaultExpiration)
	return mapping, nil
}
