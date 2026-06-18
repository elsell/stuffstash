package config

import "os"

const (
	envHTTPAddr     = "STUFF_STASH_HTTP_ADDR"
	defaultHTTPAddr = ":8080"
)

type Config struct {
	HTTPAddr string
}

func Load() Config {
	return Config{
		HTTPAddr: envOrDefault(envHTTPAddr, defaultHTTPAddr),
	}
}

func envOrDefault(name string, fallback string) string {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	return value
}
