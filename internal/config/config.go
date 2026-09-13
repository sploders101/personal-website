package config

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"github.com/spf13/viper"
)

type ErrMissingVar struct {
	Varname string
}

func (err ErrMissingVar) Error() string {
	return fmt.Sprintf("missing required variable: %v", err.Varname)
}

type ErrUnknownStorageBackend struct {
	BackendName string
}

func (err ErrUnknownStorageBackend) Error() string {
	return fmt.Sprintf("unknown storage backend: %v", err.BackendName)
}

type ServerConfig struct {
	DbConfig              DbConfig
	DefaultStorageBackend string
	StorageBackends       map[string]StorageConfig
}

type DbConfig struct {
	Dialect string
	Url     string
}

type StorageConfig interface {
	GetType() string
}

func getRequired(varname string) (string, error) {
	value := os.Getenv(varname)
	if value == "" {
		return "", ErrMissingVar{varname}
	}
	return value, nil
}

func GetServerConfig() (*ServerConfig, error) {
	config := &ServerConfig{}

	dbDialect, err := getRequired("DATABASE_DIALECT")
	if err != nil {
		return nil, err
	}
	switch dbDialect {
	case "postgres":
		config.DbConfig.Dialect = dbDialect
		dbUrl, err := getRequired("DATABASE_URL")
		if err != nil {
			return nil, err
		}
		config.DbConfig.Url = dbUrl
	}

	storageBackendsRaw, err := getRequired("STORAGE_BACKENDS")
	if err != nil {
		return nil, err
	}
	storageBackends := strings.Split(storageBackendsRaw, ",")
	defaultBackend := os.Getenv("DEFAULT_STORAGE_BACKEND")
	if defaultBackend == "" {
		if len(storageBackends) > 1 {
			return nil, ErrMissingVar{"DEFAULT_STORAGE_BACKEND"}
		}
	} else if !slices.Contains(storageBackends, defaultBackend) {
		return nil, ErrUnknownStorageBackend{defaultBackend}
	}
	for _, backendName := range storageBackends {
		backendType, err := getRequired("STORAGE_"+backendName+"_TYPE")
		if err != nil
		getS3BackendConfig(backendName)
	}
}

func getS3BackendConfig(backendName string) StorageConfig {
	
}
