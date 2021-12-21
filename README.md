# modd
[![test](https://github.com/terakoya76/modd/actions/workflows/test.yml/badge.svg)](https://github.com/terakoya76/modd/actions/workflows/test.yml)

monitor of datadog monitors

modd uniformly targets the resources belonging to the metrics monitored by Datadog monitors, and detects their leakage.

## Required
To run modd, datadog API/App keys environment variables are required.

```bash
export DD_CLIENT_API_KEY=xxxx
export DD_CLIENT_APP_KEY=yyyy
```

Also, permissions to get resources that should be monitored are required.

```bash
# for AWS
aws sts get-caller-identity
{
    "UserId": "xxxxxx:yyyyyy@example.com",
    "Account": "zzzzzz",
    "Arn": "arn:aws:sts::zzzzzz:assumed-role/my-role/xxxxxx:yyyyyy@example.com"
}
```

## Tag Matcher Configuration

In some cases, it is necessary to control in detail whether a resource that belongs to a metric is a resource that should be monitored or not.
For example, the metrics of aws rds are different by db engine.

In this case, it can be controlled by setting up rules for matching monitor tags with resource tags from environment variables.

```bash
export AWS_RDS_AWS_TAG_KEY=dbengine
export AWS_RDS_DATADOG_TAG_KEY=dbengine
```

## Support Integration

AWS
* AutoScalingGroup
* Elasticache
* Kinesis
* OpenSearch Service
* RDS
* SQS

