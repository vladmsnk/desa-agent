package config

import (
	"fmt"
	"os"
	"strconv"
)

type IdentityProviderType string

const (
	IdentityProviderTypeActiveDirectory IdentityProviderType = "active_directory"
	IdentityProviderTypeLDAP            IdentityProviderType = "ldap"
)

type Config struct {
	GRPC    GRPCConfig
	IDP     IDPConfig
	Storage StorageConfig
}

type GRPCConfig struct {
	Host string
	Port int
}

func (g GRPCConfig) Address() string {
	return fmt.Sprintf("%s:%d", g.Host, g.Port)
}

type IDPConfig struct {
	Type     IdentityProviderType
	Host     string
	Port     int
	BaseDN   string
	BindDN   string
	BindPass string
	UseTLS   bool
}

type StorageConfig struct {
	Path     string
	InMemory bool
}

func LoadFromEnv() (*Config, error) {
	cfg := &Config{
		GRPC: GRPCConfig{
			Host: getEnv("GRPC_HOST", "0.0.0.0"),
			Port: getEnvInt("GRPC_PORT", 50051),
		},
		IDP: IDPConfig{
			Type:     IdentityProviderType(getEnv("IDP_TYPE", string(IdentityProviderTypeLDAP))),
			Host:     getEnv("IDP_HOST", "localhost"),
			Port:     getEnvInt("IDP_PORT", 389),
			BaseDN:   getEnv("IDP_BASE_DN", ""),
			BindDN:   getEnv("IDP_BIND_DN", ""),
			BindPass: getEnv("IDP_BIND_PASS", ""),
			UseTLS:   getEnvBool("IDP_USE_TLS", false),
		},
		Storage: StorageConfig{
			Path:     getEnv("STORAGE_PATH", "./data"),
			InMemory: getEnvBool("STORAGE_IN_MEMORY", false),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	switch c.IDP.Type {
	case IdentityProviderTypeActiveDirectory, IdentityProviderTypeLDAP:
	default:
		return fmt.Errorf("invalid IDP_TYPE: %s, must be one of: %s, %s",
			c.IDP.Type, IdentityProviderTypeActiveDirectory, IdentityProviderTypeLDAP)
	}

	if c.IDP.Host == "" {
		return fmt.Errorf("IDP_HOST is required")
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}
