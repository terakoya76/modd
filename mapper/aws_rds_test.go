package mapper_test

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/aws/smithy-go/middleware"
	goCache "github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"

	"github.com/terakoya76/modd/mapper"
)

// dummyAwsRdsClient implements AwsRdsClient interface for faking AWS API.
type dummyAwsRdsClient struct{}

// ListTopics implements AwsRdsClient for dummyAwsRdsClient.
func (c *dummyAwsRdsClient) DescribeDBInstances(
	ctx context.Context,
	params *rds.DescribeDBInstancesInput,
	optFns ...func(*rds.Options),
) (*rds.DescribeDBInstancesOutput, error) {
	var output rds.DescribeDBInstancesOutput

	if params.Marker != nil {
		output = rds.DescribeDBInstancesOutput{
			DBInstances: []types.DBInstance{
				{
					DBInstanceIdentifier: aws.String("db10"),
					TagList: []types.Tag{
						{
							Key:   aws.String("key10"),
							Value: aws.String("val10"),
						},
						{
							Key:   aws.String("key20"),
							Value: aws.String("val20"),
						},
					},
				},
				{
					DBInstanceIdentifier: aws.String("db20"),
					TagList: []types.Tag{
						{
							Key:   aws.String("key30"),
							Value: aws.String("val30"),
						},
						{
							Key:   aws.String("key40"),
							Value: aws.String("val40"),
						},
					},
				},
			},
			Marker:         nil,
			ResultMetadata: middleware.Metadata{},
		}
	} else {
		output = rds.DescribeDBInstancesOutput{
			DBInstances: []types.DBInstance{
				{
					DBInstanceIdentifier: aws.String("db1"),
					TagList: []types.Tag{
						{
							Key:   aws.String("key1"),
							Value: aws.String("val1"),
						},
						{
							Key:   aws.String("key2"),
							Value: aws.String("val2"),
						},
					},
				},
				{
					DBInstanceIdentifier: aws.String("db2"),
					TagList: []types.Tag{
						{
							Key:   aws.String("key3"),
							Value: aws.String("val3"),
						},
						{
							Key:   aws.String("key4"),
							Value: aws.String("val4"),
						},
					},
				},
			},
			Marker:         aws.String("next marker"),
			ResultMetadata: middleware.Metadata{},
		}
	}

	return &output, nil
}

func Test_AwsRds_GetTagsMapping(t *testing.T) {
	cache := goCache.New(60*time.Minute, 10*time.Minute)

	cases := []struct {
		name     string
		expected map[string]mapper.Tags
		err      error
	}{
		{
			name: "fake test",
			expected: map[string]mapper.Tags{
				"db1":  []string{"key1:val1", "key2:val2"},
				"db2":  []string{"key3:val3", "key4:val4"},
				"db10": []string{"key10:val10", "key20:val20"},
				"db20": []string{"key30:val30", "key40:val40"},
			},
			err: nil,
		},
	}

	for _, c := range cases {
		client := dummyAwsRdsClient{}
		m := mapper.BuildAwsRdsTagsMapper(cache, &client)
		actual, err := m.GetTagsMapping(context.TODO())
		if !assert.Equal(t, c.err, err) {
			t.Errorf("case: %s is failed, expected: %+v, actual: %+v\n", c.name, c.err, err)
		}
		if !assert.Equal(t, c.expected, actual) {
			t.Errorf("case: %s is failed, expected: %+v, actual: %+v\n", c.name, c.expected, actual)
		}
	}
}
