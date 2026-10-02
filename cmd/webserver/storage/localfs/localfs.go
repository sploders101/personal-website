package localfs

import (
	"context"
	"io"
	"log/slog"
	"os"
)

type Config struct {
	Path string
}

type LocalFSStore struct {
	root *os.Root
}

func NewLocalFSStore(ctx context.Context, cfg Config) (*LocalFSStore, error) {
	dir, err := os.OpenRoot(cfg.Path)
	if err != nil {
		return nil, err
	}
	return &LocalFSStore{root: dir}, nil
}

type TruncatedFile struct {
	io.Reader
	io.Closer
}

func (store LocalFSStore) GetFile(
	ctx context.Context,
	objectName string,
) (io.ReadSeekCloser, error) {
	file, err := store.root.Open(objectName)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (store LocalFSStore) PutFile(
	ctx context.Context,
	objectName string,
	size int64,
	file io.Reader,
) error {
	localfile, err := store.root.Create(objectName)
	if err != nil {
		return err
	}
	bytesCopied, copyErr := io.Copy(localfile, io.LimitReader(file, size))
	if err := localfile.Close(); err != nil {
		slog.Error(
			"Failed to close file",
			"storageDriver",
			"localfs",
			"objectName",
			objectName,
			"error",
			err,
		)
	}
	if bytesCopied != size || copyErr != nil {
		if err := store.root.Remove(objectName); err != nil {
			slog.Error(
				"Failed to remove file after failed upload",
				"storageDriver",
				"localfs",
				"objectName",
				objectName,
				"error",
				err,
			)
		}

		if err := ctx.Err(); err != nil {
			return err
		}
		if copyErr != nil {
			return copyErr // From io.Copy
		}
		return io.ErrUnexpectedEOF
	}
	return nil
}

func (store LocalFSStore) DeleteFile(ctx context.Context, objectName string) error {
	if err := store.root.Remove(objectName); err != nil {
		return err
	}
	return nil
}
