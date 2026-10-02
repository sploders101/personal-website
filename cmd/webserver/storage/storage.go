package storage

import (
	"context"
	"io"
)

type StorageDriver interface {
	GetFile(ctx context.Context, objectName string) (io.ReadSeekCloser, error)
	PutFile(ctx context.Context, objectName string, size int64, file io.Reader) error
	DeleteFile(ctx context.Context, objectName string) error
}
