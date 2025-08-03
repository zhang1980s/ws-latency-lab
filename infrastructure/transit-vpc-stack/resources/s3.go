package resources

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/zhang1980s/ws-latency-lab/infrastructure/common/config"
	"github.com/zhang1980s/ws-latency-lab/infrastructure/common/utils"
)

// generateRandomSuffix generates a random 8-character hex string
func generateRandomSuffix() (string, error) {
	bytes := make([]byte, 4) // 4 bytes = 8 hex characters
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// CreateAccessLogBucket creates an S3 bucket for storing ALB and NLB access logs
func CreateAccessLogBucket(ctx *pulumi.Context, cfg *config.Config) (*s3.Bucket, error) {
	// Generate a random suffix for the bucket name
	randomSuffix, err := generateRandomSuffix()
	if err != nil {
		return nil, fmt.Errorf("failed to generate random suffix: %w", err)
	}

	// Create a unique bucket name using the project, environment, and a random suffix
	bucketName := fmt.Sprintf("%s-%s-lb-access-logs-%s", cfg.Project, cfg.Environment, randomSuffix)

	// Create the S3 bucket
	bucket, err := s3.NewBucket(ctx, bucketName, &s3.BucketArgs{
		Bucket: pulumi.String(bucketName),
		// Configure server-side encryption
		ServerSideEncryptionConfiguration: &s3.BucketServerSideEncryptionConfigurationArgs{
			Rule: &s3.BucketServerSideEncryptionConfigurationRuleArgs{
				ApplyServerSideEncryptionByDefault: &s3.BucketServerSideEncryptionConfigurationRuleApplyServerSideEncryptionByDefaultArgs{
					SseAlgorithm: pulumi.String("AES256"),
				},
			},
		},
		// Configure lifecycle rules for log rotation
		LifecycleRules: s3.BucketLifecycleRuleArray{
			&s3.BucketLifecycleRuleArgs{
				Enabled: pulumi.Bool(true),
				Id:      pulumi.String("expire-old-logs"),
				Expiration: &s3.BucketLifecycleRuleExpirationArgs{
					Days: pulumi.Int(90), // Expire logs after 90 days
				},
			},
		},
		// Block public access - using individual properties
		Acl:          pulumi.String("private"),
		ForceDestroy: pulumi.Bool(true),
		// Apply tags
		Tags: utils.ApplyTags(ctx, bucketName, utils.GetNamedTags(bucketName, cfg.Environment, cfg.Project, cfg.Owner, cfg.CustomTags)),
	})
	if err != nil {
		return nil, err
	}

	// Create bucket policy to allow ALB and NLB to write access logs
	// The AWS ELB service account IDs are region-specific
	// For ap-east-1 (Hong Kong), the ELB account ID is 754344448648
	elbAccountId := "754344448648" // ELB account ID for ap-east-1

	// Create bucket policy
	_, err = s3.NewBucketPolicy(ctx, fmt.Sprintf("%s-policy", bucketName), &s3.BucketPolicyArgs{
		Bucket: bucket.ID(),
		Policy: pulumi.All(bucket.Bucket).ApplyT(func(args []interface{}) (string, error) {
			bucketName := args[0].(string)
			return fmt.Sprintf(`{
				"Version": "2012-10-17",
				"Statement": [
					{
						"Effect": "Allow",
						"Principal": {
							"AWS": "arn:aws:iam::%s:root"
						},
						"Action": "s3:PutObject",
						"Resource": "arn:aws:s3:::%s/*"
					},
					{
						"Effect": "Allow",
						"Principal": {
							"Service": "delivery.logs.amazonaws.com"
						},
						"Action": "s3:PutObject",
						"Resource": "arn:aws:s3:::%s/*",
						"Condition": {
							"StringEquals": {
								"s3:x-amz-acl": "bucket-owner-full-control"
							}
						}
					},
					{
						"Effect": "Allow",
						"Principal": {
							"Service": "delivery.logs.amazonaws.com"
						},
						"Action": "s3:GetBucketAcl",
						"Resource": "arn:aws:s3:::%s"
					}
				]
			}`, elbAccountId, bucketName, bucketName, bucketName), nil
		}).(pulumi.StringOutput),
	})
	if err != nil {
		return nil, err
	}

	return bucket, nil
}
