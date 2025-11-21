package repository

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
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

// GetPresignedPutURL generates a presigned URL for uploading a file
func (s *S3AssetRepository) GetPresignedPutURL(ctx context.Context, key string, contentType string, duration time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s.client)

	request, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = duration
	})
	if err != nil {
		s.logger.Error("Failed to generate presigned PUT URL",
			zap.String("bucket", s.bucketName),
			zap.String("key", key),
			zap.String("content_type", contentType),
			zap.Error(err),
		)
		return "", fmt.Errorf("failed to generate presigned PUT URL: %w", err)
	}

	s.logger.Info("Presigned PUT URL generated",
		zap.String("key", key),
		zap.String("content_type", contentType),
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

// DeleteFile deletes a file from S3
func (s *S3AssetRepository) DeleteFile(ctx context.Context, bucket, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		s.logger.Error("Failed to delete file from S3",
			zap.String("bucket", bucket),
			zap.String("key", key),
			zap.Error(err),
		)
		return fmt.Errorf("failed to delete file from S3: %w", err)
	}

	s.logger.Info("File deleted from S3",
		zap.String("bucket", bucket),
		zap.String("key", key),
	)
	return nil
}

// DeletePrefix deletes all objects under a given prefix (best-effort cleanup).
func (s *S3AssetRepository) DeletePrefix(ctx context.Context, bucket, prefix string) error {
	paginator := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list objects for prefix %s: %w", prefix, err)
		}

		if len(page.Contents) == 0 {
			continue
		}

		identifiers := make([]types.ObjectIdentifier, 0, len(page.Contents))
		for _, object := range page.Contents {
			if object.Key == nil {
				continue
			}
			identifiers = append(identifiers, types.ObjectIdentifier{Key: object.Key})
		}

		if len(identifiers) == 0 {
			continue
		}

		_, err = s.client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(bucket),
			Delete: &types.Delete{
				Objects: identifiers,
				Quiet:   aws.Bool(true),
			},
		})
		if err != nil {
			return fmt.Errorf("failed to delete objects for prefix %s: %w", prefix, err)
		}
	}

	s.logger.Info("Deleted S3 assets for prefix",
		zap.String("bucket", bucket),
		zap.String("prefix", prefix),
	)
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
