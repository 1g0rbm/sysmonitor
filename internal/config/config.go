package config

import (
	"flag"
	"os"
	"strconv"
	"time"
)

const (
	defaultAddress        = "127.0.0.1:8080"
	defaultReportInterval = 10 * time.Second
	defaultPollInterval   = 2 * time.Second
	defaultStoreInterval  = 300 * time.Second
	defaultStoreFile      = "/tmp/devops-metrics-db.json"
	defaultRestore        = true
)

var (
	address        string
	reportInterval time.Duration
	pollInterval   time.Duration
	storeInterval  time.Duration
	storeFile      string
	restore        bool
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

func GetConfigServer() *ServerConfig {
	flag.StringVar(&address, "a", defaultAddress, "-a=<VALUE>")
	flag.DurationVar(&storeInterval, "i", defaultStoreInterval, "-i=<VALUE>")
	flag.StringVar(&storeFile, "f", defaultStoreFile, "-f=<VALUE")
	flag.BoolVar(&restore, "r", defaultRestore, "-r=<VALUE>")

	flag.Parse()

	return &ServerConfig{
		Address:       getEnvString("ADDRESS", address),
		StoreInterval: getEnvDuration("STORE_INTERVAL", storeInterval),
		StoreFile:     getEnvString("STORE_FILE", storeFile),
		Restore:       getEnvBool("RESTORE", restore),
	}
}

func GetConfigAgent() *AgentConfig {
	flag.StringVar(&address, "a", defaultAddress, "-a=<VALUE>")
	flag.DurationVar(&reportInterval, "r", defaultReportInterval, "-r=<VALUE>")
	flag.DurationVar(&pollInterval, "p", defaultPollInterval, "-p=<VALUE>")

	flag.Parse()

	return &AgentConfig{
		Address:        getEnvString("ADDRESS", address),
		ReportInterval: getEnvDuration("REPORT_INTERVAL", reportInterval),
		PollInterval:   getEnvDuration("POLL_INTERVAL", pollInterval),
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
