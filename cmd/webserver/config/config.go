package config

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

const (
	secretEnvSuffix  = "_env"
	secretFileSuffix = "_file"
)

type ServerConfig struct {
	BaseUrl        string               `json:"base_url"`
	Authentication AuthenticationConfig `json:"authentication"`
	Database       DatabaseConfig       `json:"database"`
	Storage        StorageBackends      `json:"storage"`
	Secrets        SecretConfig         `json:"secrets"`
}

type DatabaseConfig struct {
	Dialect string `json:"dialect"`
	URL     string `json:"url"`
}

type SecretConfig struct {
	CsrfSecret string `json:"csrf_secret"`
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

	if cfg.BaseUrl == "" {
		slog.Warn("Missing base_url from config. Some features may not work properly.")
	}
	if cfg.Authentication.Local.Enabled {
		slog.Warn("Local authentication not yet implemented. Please use OIDC.")
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
