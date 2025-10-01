package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type S3Client struct {
	client *s3.Client
	bucket string
}

// NewS3Client initializes an S3 client bound to a specific bucket
func NewS3Client(ctx context.Context, bucket string) (*S3Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-west-2"),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg)
	return &S3Client{
		client: client,
		bucket: bucket,
	}, nil
}

// UploadFile uploads a single file at localPath to the given key in S3
func (s *S3Client) UploadFile(ctx context.Context, key string, localPath string) error {
	f, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   f,
		ACL:    types.ObjectCannedACLPrivate, // keep objects private, accessed only by signed URLs
	})
	if err != nil {
		return err
	}

	fmt.Println("Uploaded:", key)
	return nil
}

// UploadFolder recursively uploads an entire local folder (e.g. processed HLS outputs)
func (s *S3Client) UploadFolder(ctx context.Context, localDir, prefix string) error {
	return filepath.Walk(localDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(localDir, path)
		if err != nil {
			return err
		}

		key := filepath.ToSlash(filepath.Join(prefix, relPath))
		err = s.UploadFile(ctx, key, path)
		if err != nil {
			return fmt.Errorf("failed to upload %s: %w", path, err)
		}
		return nil
	})
}

// GenerateSignedURL creates a time-limited signed URL for an object in S3
func (s *S3Client) GenerateSignedURL(ctx context.Context, key string, duration time.Duration) (string, error) {
	presigner := s3.NewPresignClient(s.client)

	presignOutput, err := presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = duration
	})
	if err != nil {
		return "", err
	}

	return presignOutput.URL, nil
}
