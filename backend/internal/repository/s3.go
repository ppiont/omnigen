package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"
)

// S3Service handles S3 operations
type S3Service struct {
	client     *s3.Client
	bucketName string
	logger     *zap.Logger
}

// NewS3Service creates a new S3 service
func NewS3Service(
	client *s3.Client,
	bucketName string,
	logger *zap.Logger,
) *S3Service {
	return &S3Service{
		client:     client,
		bucketName: bucketName,
		logger:     logger,
	}
}

// GetPresignedURL generates a presigned URL for downloading a video
func (s *S3Service) GetPresignedURL(ctx context.Context, key string, duration time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s.client)

	request, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = duration
	})
	if err != nil {
		s.logger.Error("Failed to generate presigned URL",
			zap.String("bucket", s.bucketName),
			zap.String("key", key),
			zap.Error(err),
		)
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	s.logger.Info("Presigned URL generated",
		zap.String("key", key),
		zap.Duration("expiration", duration),
	)

	return request.URL, nil
}

// DeleteObject deletes an object from S3
func (s *S3Service) DeleteObject(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		s.logger.Error("Failed to delete S3 object",
			zap.String("bucket", s.bucketName),
			zap.String("key", key),
			zap.Error(err),
		)
		return fmt.Errorf("failed to delete S3 object: %w", err)
	}

	s.logger.Info("S3 object deleted successfully", zap.String("key", key))
	return nil
}

// HealthCheck performs a lightweight health check on S3
func (s *S3Service) HealthCheck(ctx context.Context) error {
	_, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(s.bucketName),
	})
	if err != nil {
		return fmt.Errorf("s3 health check failed: %w", err)
	}
	return nil
}
