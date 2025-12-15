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

// Initializes an S3 client bound to a specific bucket
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

// Uploads a single file at localPath to the given key in S3
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
		ACL:    types.ObjectCannedACLPrivate,
	})
	if err != nil {
		return err
	}

	fmt.Println("Uploaded:", key)
	return nil
}

// Recursively uploads an entire local folder
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

// Creates a time-limited signed URL for an S3 obj
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



func (s *S3Client) ListObjects(ctx context.Context) ([]string, error) {
	paginator := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
	})

	var keys []string
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", err)
		}

		for _, obj := range page.Contents {
			keys = append(keys, *obj.Key)
		}
	}

	return keys, nil
}

func (s *S3Client) DeleteAllObjects(ctx context.Context) error {
	// List all objects in the bucket
	paginator := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
	})

	var deletedCount int
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list objects: %w", err)
		}

		if len(page.Contents) == 0 {
			continue
		}

		var objectIdentifiers []types.ObjectIdentifier
		for _, obj := range page.Contents {
			objectIdentifiers = append(objectIdentifiers, types.ObjectIdentifier{
				Key: obj.Key,
			})
		}

		_, err = s.client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(s.bucket),
			Delete: &types.Delete{
				Objects: objectIdentifiers,
				Quiet:   aws.Bool(true),
			},
		})
		if err != nil {
			return fmt.Errorf("failed to delete objects: %w", err)
		}

		deletedCount += len(objectIdentifiers)
		fmt.Println("Deleted", deletedCount, "objects from bucket", s.bucket)
	}

	if deletedCount == 0 {
		fmt.Println("No objects found in bucket", s.bucket)
	} else {
		fmt.Println("Successfully deleted all", deletedCount, "objects from bucket", s.bucket)
	}

	return nil
}
