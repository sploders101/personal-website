package main

import (
	"context"
	"errors"

	"github.com/sploders101/personal-website/cmd/webserver/config"
	"github.com/sploders101/personal-website/cmd/webserver/storage"
	"github.com/sploders101/personal-website/cmd/webserver/storage/localfs"
	"github.com/sploders101/personal-website/cmd/webserver/storage/s3"
)

func createStorageDriver(
	ctx context.Context,
	cfg config.ServerConfig,
) (storage.StorageDriver, error) {
	switch {
	case cfg.Storage.LocalFS != nil:
		return localfs.NewLocalFSStore(ctx, localfs.Config{
			Path: cfg.Storage.LocalFS.Path,
		})
	case cfg.Storage.S3 != nil:
		return s3.NewS3Store(ctx, s3.Config{
			Bucket:    cfg.Storage.S3.BucketName,
			Endpoint:  cfg.Storage.S3.Endpoint,
			Region:    cfg.Storage.S3.BucketRegion,
			AccessKey: cfg.Storage.S3.AccessKeyID,
			SecretKey: cfg.Storage.S3.AccessKeySecret,
			UseSSL:    cfg.Storage.S3.UseSSL,
		})
	default:
		return nil, errors.New("no storage provider configured")
	}
}
