package storage

import (
	"context"
	"io"
)

type GetFileOptions struct {
	RequestRange bool
	Start        int64
	End          int64
}

type StorageDriver interface {
	GetFile(ctx context.Context, objectName string, opts GetFileOptions) (io.ReadCloser, error)
	PutFile(ctx context.Context, objectName string, size int64, file io.Reader) error
	DeleteFile(ctx context.Context, objectName string) error
}
