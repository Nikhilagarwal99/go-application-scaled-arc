package utils

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	appconfig "github.com/nikhilAgarwal99/go-application-scaled-arc/internal/config"
)

// S3Client implements StorageProvider using AWS S3.
// Swap this with GCSClient or R2Client tomorrow — nothing else changes.
type S3Client struct {
	client    *s3.Client
	bucket    string
	bucketURL string
}

// Compile time check — fails immediately at build if S3Client
// ever stops implementing StorageProvider
var _ StorageProvider = (*S3Client)(nil)

func NewS3Client(cfg *appconfig.Config) (*S3Client, error) {
	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(cfg.AWSRegion),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				cfg.AWSAccessKey,
				cfg.AWSSecretKey,
				"",
			),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &S3Client{
		client:    s3.NewFromConfig(awsCfg),
		bucket:    cfg.AWSBucket,
		bucketURL: fmt.Sprintf("https://%s.s3.%s.amazonaws.com", cfg.AWSBucket, cfg.AWSRegion),
	}, nil
}

// upload implements StorageProvider.
// Lowercase — intentionally unexported.
// Nobody outside this package calls upload() directly.
// Everything goes through UploadFile() in storage.go.
func (s *S3Client) upload(
	ctx context.Context,
	file multipart.File,
	header *multipart.FileHeader,
	folder string,
	resultChan chan<- UploadResult,
) {
	buf := bytes.NewBuffer(nil)
	if _, err := buf.ReadFrom(file); err != nil {
		resultChan <- UploadResult{Err: fmt.Errorf("failed to read file: %w", err)}
		return
	}

	ext := filepath.Ext(header.Filename)
	key := fmt.Sprintf("%s/%s%s", folder, uuid.New().String(), ext)

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(buf.Bytes()),
		ContentType: aws.String(contentTypeFromExt(ext)),
	})
	if err != nil {
		resultChan <- UploadResult{Err: fmt.Errorf("s3 put failed: %w", err)}
		return
	}

	resultChan <- UploadResult{URL: fmt.Sprintf("%s/%s", s.bucketURL, key)}
}

func contentTypeFromExt(ext string) string {
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}
