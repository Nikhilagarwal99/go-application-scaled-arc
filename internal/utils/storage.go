package utils

import (
	"context"
	"mime/multipart"

	"github.com/nikhilAgarwal99/go-application-scaled-arc/internal/logger"
	"go.uber.org/zap"
)

type UploadResult struct {
	URL string
	Err error
}

/*
StorageProvider is the interface every cloud storage implementation must satisfy.
S3, GCS, Cloudflare R2 — all implement this one method.
The channel design lets it run inside a goroutine without blocking. */

type StorageProvider interface {
	upload(ctx context.Context, file multipart.File, header *multipart.FileHeader, folder string, result chan<- UploadResult)
}

/*
UploadFile is the only function the rest of the app ever calls for file uploads.
It hides goroutines, channels, and cloud provider details completely.

One line usage from any service:

	url, err := utils.UploadFile(ctx, s.storage, file, header, "avatars") */

func UploadFile(ctx context.Context, provider StorageProvider, file multipart.File, header *multipart.FileHeader, folder string) (string, error) {
	resultChan := make(chan UploadResult, 1) // buffered — goroutine never leaks

	go provider.upload(ctx, file, header, folder, resultChan)

	select {
	case result := <-resultChan:
		if result.Err != nil {
			logger.Error("file upload failed",
				zap.String("provider", folder),
				zap.String("filename", header.Filename),
				zap.Error(result.Err),
			)
			return "", result.Err
		}
		logger.Info("file uploaded successfully",
			zap.String("url", result.URL),
			zap.String("folder", folder),
		)
		return result.URL, nil

	case <-ctx.Done():
		logger.Warn("file upload aborted — context cancelled",
			zap.String("folder", folder),
			zap.String("filename", header.Filename),
		)
		return "", ctx.Err()
	}
}
