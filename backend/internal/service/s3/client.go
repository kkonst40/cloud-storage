package s3

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kkonst40/cloud-storage/backend/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func NewClient(cfg *config.Config) (*minio.Client, error) {
	client, err := minio.New(cfg.S3Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.S3AccessKey, cfg.S3SecretAccess, ""),
		Secure: cfg.S3UseSSL,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to init minio client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	exists, err := client.BucketExists(ctx, cfg.S3Bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		log.Printf("Bucket %s does not exist. Creating...", cfg.S3Bucket)
		err = client.MakeBucket(ctx, cfg.S3Bucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
		log.Printf("Bucket %s created successfully", cfg.S3Bucket)
	}

	return client, nil
}
