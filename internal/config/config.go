package config

import (
	"os"
	"strconv"
	"time"
)

const (
	Address        = "127.0.0.1:8080"
	ReportInterval = 10 * time.Second
	PollInterval   = 2 * time.Second
	StoreInterval  = 300 * time.Second
	StoreFile      = "/tmp/devops-metrics-db.json"
	Restore        = true
)

type ServerConfig struct {
	Address       string
	StoreInterval time.Duration
	StoreFile     string
	Restore       bool
}

type AgentConfig struct {
	Address        string
	ReportInterval time.Duration
	PollInterval   time.Duration
}

func GetConfigServer() ServerConfig {
	return ServerConfig{
		Address:       getEnvString("ADDRESS", Address),
		StoreInterval: getEnvDuration("STORE_INTERVAL", StoreInterval),
		StoreFile:     getEnvString("STORE_FILE", StoreFile),
		Restore:       getEnvBool("RESTORE", Restore),
	}
}

func GetConfigAgent() AgentConfig {
	return AgentConfig{
		Address:        getEnvString("ADDRESS", Address),
		ReportInterval: getEnvDuration("REPORT_INTERVAL", ReportInterval),
		PollInterval:   getEnvDuration("POLL_INTERVAL", PollInterval),
	}
}

func getEnvString(name string, defaultValue string) string {
	value, ok := os.LookupEnv(name)
	if !ok {
		return defaultValue
	}

	return value
}

func getEnvDuration(name string, defaultValue time.Duration) time.Duration {
	value, ok := os.LookupEnv(name)
	if !ok {
		return defaultValue
	}

	normalizedValue, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue
	}

	return normalizedValue
}

func getEnvBool(name string, defaultValue bool) bool {
	value, ok := os.LookupEnv(name)
	if !ok {
		return defaultValue
	}

	b, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	return b
}
