package app

import (
	"backend/internal/infrastructure/repositories/minio"
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func initS3(cfg minio.Config) (*s3.Client, error) {
	awsCfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				cfg.AccessKey,
				cfg.SecretKey,
				"",
			),
		),
		config.WithRegion(cfg.Region),
	)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.EndpointResolver = s3.EndpointResolverFromURL(cfg.Endpoint)
		o.UsePathStyle = true
	})

	if err := ensureBucket(context.Background(), client, cfg.Bucket); err != nil {
		return nil, fmt.Errorf("ensure bucket: %w", err)
	}

	return client, nil
}

func ensureBucket(ctx context.Context, client *s3.Client, bucket string) error {
	_, err := client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})

	if err == nil {
		return nil
	}

	_, err = client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		return fmt.Errorf("create bucket %s: %w", bucket, err)
	}

	return nil
}
