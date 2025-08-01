package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var s3Client *s3.Client

// initS3Client initializes the S3 client using default credentials chain.
func initS3Client(ctx context.Context) error {
	if s3Client != nil {
		return nil
	}
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}
	s3Client = s3.NewFromConfig(cfg)
	return nil
}

// downloadS3File downloads a file from S3 to a local path.
func downloadS3File(ctx context.Context, s3Path, localSaveDir string) (string, error) {
	if err := initS3Client(ctx); err != nil {
		return "", err
	}

	bucket, key, err := parseS3Path(s3Path)
	if err != nil {
		return "", err
	}

	result, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get object %s from bucket %s: %w", key, bucket, err)
	}
	defer result.Body.Close()

	localFilePath := filepath.Join(localSaveDir, filepath.Base(key))
	outFile, err := os.Create(localFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to create local file %s: %w", localFilePath, err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, result.Body)
	if err != nil {
		return "", fmt.Errorf("failed to copy S3 object to local file: %w", err)
	}

	return localFilePath, nil
}

// parseS3Path splits an s3:// path into bucket and key.
func parseS3Path(s3Path string) (bucket, key string, err error) {
	if !strings.HasPrefix(s3Path, "s3://") {
		return "", "", fmt.Errorf("invalid S3 path: must start with s3://")
	}
	pathWithoutPrefix := strings.TrimPrefix(s3Path, "s3://")
	parts := strings.SplitN(pathWithoutPrefix, "/", 2)
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid S3 path format: %s", s3Path)
	}
	return parts[0], parts[1], nil
}
