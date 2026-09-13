package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const (
	secretEnvSuffix  = "_env"
	secretFileSuffix = "_file"
)

type ServerConfig struct {
	Database DatabaseConfig `json:"database"`
	Storage  Backends       `json:"storage"`
}

type DatabaseConfig struct {
	Dialect string `json:"dialect"`
	URL     string `json:"url"`
}

// Backends maps each storage backend name to its configuration.
type Backends map[string]Backend

// UnmarshalJSON decodes each storage entry into the concrete backend type
// selected by its "type" field.
func (m *Backends) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*m = make(Backends, len(raw))
	for name, val := range raw {
		backend, err := decodeBackend(val)
		if err != nil {
			return fmt.Errorf("storage.%s: %w", name, err)
		}
		(*m)[name] = backend
	}
	return nil
}

// Backend is implemented by every storage backend configuration. The "type"
// field in the JSON file selects which concrete backend is decoded.
type Backend interface {
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
func decodeBackend(data []byte) (Backend, error) {
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

// Load reads the JSON config file at path, resolves any secret references,
// and decodes the result into a ServerConfig.
func Load(path string) (ServerConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ServerConfig{}, fmt.Errorf("read config file %s: %w", path, err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return ServerConfig{}, fmt.Errorf("parse config file %s: %w", path, err)
	}
	if err := resolveSecrets(m); err != nil {
		return ServerConfig{}, err
	}
	normalized, err := json.Marshal(m)
	if err != nil {
		return ServerConfig{}, fmt.Errorf("normalize config: %w", err)
	}
	var cfg ServerConfig
	if err := json.Unmarshal(normalized, &cfg); err != nil {
		return ServerConfig{}, fmt.Errorf("decode config file %s: %w", path, err)
	}
	return cfg, nil
}

// resolveSecrets walks the config tree, replacing keys ending in _env or
// _file with the value of the referenced environment variable or file, and
// rejects conflicting definitions of the same field.
func resolveSecrets(m map[string]any) error {
	for key := range m {
		if err := checkSecretConflict(m, key); err != nil {
			return err
		}
	}
	for key, val := range m {
		switch {
		case strings.HasSuffix(key, secretEnvSuffix):
			base := strings.TrimSuffix(key, secretEnvSuffix)
			name, ok := val.(string)
			if !ok || name == "" {
				return fmt.Errorf("%q must be a non-empty string (environment variable name)", key)
			}
			resolved, ok := os.LookupEnv(name)
			if !ok {
				return fmt.Errorf("environment variable %q for field %q is not set", name, base)
			}
			m[base] = resolved
			delete(m, key)
		case strings.HasSuffix(key, secretFileSuffix):
			base := strings.TrimSuffix(key, secretFileSuffix)
			path, ok := val.(string)
			if !ok || path == "" {
				return fmt.Errorf("%q must be a non-empty string (path to a secret file)", key)
			}
			contents, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("read secret file %q for field %q: %w", path, base, err)
			}
			m[base] = strings.TrimRight(string(contents), "\r\n")
			delete(m, key)
		default:
			if err := resolveNested(val); err != nil {
				return err
			}
		}
	}
	return nil
}

// resolveNested descends into child maps and slices so secret resolution
// applies at every level of the config tree.
func resolveNested(val any) error {
	switch v := val.(type) {
	case map[string]any:
		return resolveSecrets(v)
	case []any:
		for _, item := range v {
			if err := resolveNested(item); err != nil {
				return err
			}
		}
	}
	return nil
}

// checkSecretConflict returns an error if key (a _env or _file field) is
// defined alongside a literal value or the other secret form of the same field.
func checkSecretConflict(m map[string]any, key string) error {
	var base string
	isEnv := strings.HasSuffix(key, secretEnvSuffix)
	switch {
	case isEnv:
		base = strings.TrimSuffix(key, secretEnvSuffix)
	case strings.HasSuffix(key, secretFileSuffix):
		base = strings.TrimSuffix(key, secretFileSuffix)
	default:
		return nil
	}
	if _, ok := m[base]; ok {
		return fmt.Errorf("field %q is set both literally and via %q; use only one", base, key)
	}
	other := base + secretFileSuffix
	if isEnv {
		other = base + secretFileSuffix
	} else {
		other = base + secretEnvSuffix
	}
	if _, ok := m[other]; ok {
		return fmt.Errorf("field %q is set via both %q and %q; use only one", base, key, other)
	}
	return nil
}
