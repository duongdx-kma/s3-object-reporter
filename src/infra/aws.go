package infra

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// Get AWS credentials through STS Assume Role
func AssumeRole(ctx context.Context, roleArn string) (aws.Config, error) {
	awsConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return aws.Config{}, fmt.Errorf("unable to load SDK config: %w", err)
	}

	stsClient := sts.NewFromConfig(awsConfig)
	creds := stscreds.NewAssumeRoleProvider(stsClient, roleArn)

	awsConfig.Credentials = aws.NewCredentialsCache(creds)

	return awsConfig, nil
}

// ListS3Prefixes retrieves a list of unique prefixes (folders) in an S3 bucket with a given delimiter.
func ListS3Prefixes(ctx context.Context, awsConfig aws.Config, bucketName, prefix, delimiter string) ([]string, error) {
	s3Client := s3.NewFromConfig(awsConfig)

	input := &s3.ListObjectsV2Input{
		Bucket:    aws.String(bucketName),
		Prefix:    aws.String(prefix),
		Delimiter: aws.String(delimiter),
	}

	var prefixes []string

	paginator := s3.NewListObjectsV2Paginator(s3Client, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, commonPrefix := range page.CommonPrefixes {
			prefixes = append(prefixes, *commonPrefix.Prefix)
		}
	}

	return prefixes, nil
}

// ListS3ObjectsLimited retrieves objects and folders in an S3 bucket under a given prefix with a limited depth.
func ListS3ObjectsLimited(ctx context.Context, awsConfig aws.Config, bucketName, prefix, delimiter string) ([]string, error) {
	s3Client := s3.NewFromConfig(awsConfig)

	input := &s3.ListObjectsV2Input{
		Bucket:    aws.String(bucketName),
		Prefix:    aws.String(prefix),
		Delimiter: aws.String(delimiter), // Giới hạn depth để chỉ lấy object và folder ngay dưới prefix
	}

	var objects []string

	paginator := s3.NewListObjectsV2Paginator(s3Client, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		// Thêm các file trực tiếp dưới prefix
		for _, obj := range page.Contents {
			if *obj.Key != prefix { // Loại bỏ chính prefix đó
				objects = append(objects, *obj.Key)
			}
		}

		// Thêm các folder dưới prefix
		for _, commonPrefix := range page.CommonPrefixes {
			if *commonPrefix.Prefix != prefix { // Loại bỏ chính prefix đó
				objects = append(objects, *commonPrefix.Prefix)
			}
		}
	}

	return objects, nil
}
