package config

import (
	"encoding/json"
	"fmt"
)

// StorageBackends maps each storage backend name to its configuration.
type StorageBackends map[string]StorageBackend

// UnmarshalJSON decodes each storage entry into the concrete backend type
// selected by its "type" field.
func (m *StorageBackends) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*m = make(StorageBackends, len(raw))
	for name, val := range raw {
		backend, err := decodeBackend(val)
		if err != nil {
			return fmt.Errorf("storage.%s: %w", name, err)
		}
		(*m)[name] = backend
	}
	return nil
}

// StorageBackend is implemented by every storage backend configuration. The "type"
// field in the JSON file selects which concrete backend is decoded.
type StorageBackend interface {
	Kind() string
}

// S3Backend configures an S3-compatible storage backend.
type S3Backend struct {
	Type            string `json:"type"`
	Endpoint        string `json:"endpoint"`
	BucketName      string `json:"bucket_name"`
	BucketPort      int    `json:"bucket_port"`
	BucketRegion    string `json:"bucket_region"`
	BucketSubregion string `json:"bucket_subregion"`
	AccessKeyID     string `json:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret"`
}

func (b S3Backend) Kind() string { return b.Type }

// FSBackend configures a storage backend backed by a local filesystem path.
type FSBackend struct {
	Type string `json:"type"`
	Path string `json:"path"`
}

func (b FSBackend) Kind() string { return b.Type }

// decodeBackend decodes a single storage entry based on its "type" field.
func decodeBackend(data []byte) (StorageBackend, error) {
	var probe struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		return nil, err
	}
	switch probe.Type {
	case "s3":
		var backend S3Backend
		if err := json.Unmarshal(data, &backend); err != nil {
			return nil, err
		}
		return backend, nil
	case "fs":
		var backend FSBackend
		if err := json.Unmarshal(data, &backend); err != nil {
			return nil, err
		}
		return backend, nil
	default:
		return nil, fmt.Errorf("unknown storage backend type %q", probe.Type)
	}
}
