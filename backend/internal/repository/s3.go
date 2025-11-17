package repository

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"
)

// S3Service handles S3 operations
type S3AssetRepository struct {
	client     *s3.Client
	bucketName string
	logger     *zap.Logger
}

// NewS3Service creates a new S3 service
func NewS3Service(
	client *s3.Client,
	bucketName string,
	logger *zap.Logger,
) *S3AssetRepository {
	return &S3AssetRepository{
		client:     client,
		bucketName: bucketName,
		logger:     logger,
	}
}

// GetPresignedURL generates a presigned URL for downloading a video
func (s *S3AssetRepository) GetPresignedURL(ctx context.Context, key string, duration time.Duration) (string, error) {
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

// UploadFile uploads a file to S3
func (s *S3AssetRepository) UploadFile(ctx context.Context, bucket, key, filePath string, contentType string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	// Return S3 URL
	url := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucket, key)
	return url, nil
}

// DownloadFile downloads a file from S3
func (s *S3AssetRepository) DownloadFile(ctx context.Context, bucket, key, destPath string) error {
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to get object: %w", err)
	}
	defer result.Body.Close()

	file, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, result.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// HealthCheck performs a lightweight health check on S3
func (s *S3AssetRepository) HealthCheck(ctx context.Context) error {
	_, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(s.bucketName),
	})
	if err != nil {
		return fmt.Errorf("s3 health check failed: %w", err)
	}
	return nil
}
