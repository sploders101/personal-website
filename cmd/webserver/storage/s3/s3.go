package s3

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Config struct {
	Bucket    string
	Endpoint  string
	Region    string
	AccessKey string
	SecretKey string
	UseSSL    bool
}

type S3Store struct {
	bucketName string
	client     *minio.Client
}

func NewS3Store(ctx context.Context, cfg Config) (*S3Store, error) {
	cleanedEndpoint := cfg.Endpoint
	cleanedEndpoint = strings.TrimPrefix(cleanedEndpoint, "http://")
	if strings.HasPrefix(cleanedEndpoint, "https://") {
		// Prevent accidental loss of encryption
		cfg.UseSSL = true
		cleanedEndpoint = cleanedEndpoint[8:]
	}

	client, err := minio.New(cleanedEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to init storage client: %w", err)
	}

	return &S3Store{bucketName: cfg.Bucket, client: client}, nil
}

func (store S3Store) GetFile(
	ctx context.Context,
	objectName string,
) (io.ReadSeekCloser, error) {
	var mioOpts minio.GetObjectOptions
	obj, err := store.client.GetObject(ctx, store.bucketName, objectName, mioOpts)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func (store S3Store) PutFile(
	ctx context.Context,
	objectName string,
	size int64,
	file io.Reader,
) error {
	var mioOpts minio.PutObjectOptions
	_, err := store.client.PutObject(
		ctx,
		store.bucketName,
		objectName,
		file,
		size,
		mioOpts,
	)
	if err != nil {
		return err
	}
	return nil
}

func (store S3Store) DeleteFile(ctx context.Context, objectName string) error {
	var mioOpts minio.RemoveObjectOptions
	if err := store.client.RemoveObject(ctx, store.bucketName, objectName, mioOpts); err != nil {
		return err
	}
	return nil
}
