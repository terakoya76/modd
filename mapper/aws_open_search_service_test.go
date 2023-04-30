package mapper_test

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticsearchservice"
	"github.com/aws/aws-sdk-go-v2/service/elasticsearchservice/types"
	"github.com/aws/smithy-go/middleware"
	goCache "github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"

	"github.com/terakoya76/modd/mapper"
)

// dummyAwsOpenSearchServiceClient implements AwsOpenSearchServiceClient interface for faking AWS API.
type dummyAwsOpenSearchServiceClient struct{}

// ListDomainNames implements AwsOpenSearchServiceClient for dummyAwsOpenSearchServiceClient.
func (c *dummyAwsOpenSearchServiceClient) ListDomainNames(
	_ context.Context,
	_ *elasticsearchservice.ListDomainNamesInput,
	_ ...func(*elasticsearchservice.Options),
) (*elasticsearchservice.ListDomainNamesOutput, error) {
	output := elasticsearchservice.ListDomainNamesOutput{
		DomainNames: []types.DomainInfo{
			{
				DomainName: aws.String("domain1"),
				EngineType: types.EngineTypeOpenSearch,
			},
			{
				DomainName: aws.String("domain2"),
				EngineType: types.EngineTypeOpenSearch,
			},
		},
		ResultMetadata: middleware.Metadata{},
	}

	return &output, nil
}

// DescribeElasticsearchDomains implements AwsOpenSearchServiceClient for dummyAwsOpenSearchServiceClient.
func (c *dummyAwsOpenSearchServiceClient) DescribeElasticsearchDomains(
	_ context.Context,
	_ *elasticsearchservice.DescribeElasticsearchDomainsInput,
	_ ...func(*elasticsearchservice.Options),
) (*elasticsearchservice.DescribeElasticsearchDomainsOutput, error) {
	output := elasticsearchservice.DescribeElasticsearchDomainsOutput{
		DomainStatusList: []types.ElasticsearchDomainStatus{
			{
				ARN:        aws.String("domain arn 1"),
				DomainName: aws.String("domain1"),
			},
			{
				ARN:        aws.String("domain arn 2"),
				DomainName: aws.String("domain2"),
			},
		},
		ResultMetadata: middleware.Metadata{},
	}

	return &output, nil
}

// ListTags implements AwsOpenSearchServiceClient for dummyAwsOpenSearchServiceClient.
func (c *dummyAwsOpenSearchServiceClient) ListTags(
	_ context.Context,
	_ *elasticsearchservice.ListTagsInput,
	_ ...func(*elasticsearchservice.Options),
) (*elasticsearchservice.ListTagsOutput, error) {
	output := elasticsearchservice.ListTagsOutput{
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
		ResultMetadata: middleware.Metadata{},
	}

	return &output, nil
}

func Test_AwsOpenSearchService_GetTagsMapping(t *testing.T) {
	cache := goCache.New(60*time.Minute, 10*time.Minute)

	cases := []struct {
		name     string
		expected map[string]mapper.Tags
		err      error
	}{
		{
			name: "fake test",
			expected: map[string]mapper.Tags{
				"domain1": []string{"key1:val1", "key2:val2"},
				"domain2": []string{"key1:val1", "key2:val2"},
			},
			err: nil,
		},
	}

	for _, c := range cases {
		client := dummyAwsOpenSearchServiceClient{}
		m := mapper.BuildAwsOpenSearchServiceTagsMapper(cache, &client)
		actual, err := m.GetTagsMapping(context.TODO())
		if !assert.Equal(t, c.err, err) {
			t.Errorf("case: %s is failed, expected: %+v, actual: %+v\n", c.name, c.err, err)
		}
		if !assert.Equal(t, c.expected, actual) {
			t.Errorf("case: %s is failed, expected: %+v, actual: %+v\n", c.name, c.expected, actual)
		}
	}
}
