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
	defaultKey            = ""
	defaultDBDsn          = ""
)

var (
	address        string
	reportInterval time.Duration
	pollInterval   time.Duration
	storeInterval  time.Duration
	storeFile      string
	restore        bool
	key            string
	DBDsn          string
)

type ServerConfig struct {
	Address       string
	StoreInterval time.Duration
	StoreFile     string
	Restore       bool
	Key           string
	DBDsn         string
}

type AgentConfig struct {
	Address        string
	ReportInterval time.Duration
	PollInterval   time.Duration
	Key            string
}

func GetConfigServer() *ServerConfig {
	flag.StringVar(&address, "a", defaultAddress, "-a=<VALUE>")
	flag.DurationVar(&storeInterval, "i", defaultStoreInterval, "-i=<VALUE>")
	flag.StringVar(&storeFile, "f", defaultStoreFile, "-f=<VALUE")
	flag.BoolVar(&restore, "r", defaultRestore, "-r=<VALUE>")
	flag.StringVar(&key, "k", defaultKey, "-k=<KEY>")
	flag.StringVar(&DBDsn, "d", defaultDBDsn, "-d=<DATABASE_DSN>")

	flag.Parse()

	return &ServerConfig{
		Address:       getEnvString("ADDRESS", address),
		StoreInterval: getEnvDuration("STORE_INTERVAL", storeInterval),
		StoreFile:     getEnvString("STORE_FILE", storeFile),
		Restore:       getEnvBool("RESTORE", restore),
		Key:           getEnvString("KEY", key),
		DBDsn:         getEnvString("DATABASE_DSN", DBDsn),
	}
}

func (sc ServerConfig) NeedRestore() bool {
	return sc.DBDsn == "" && sc.Restore
}

func (sc ServerConfig) NeedPeriodicalStore() bool {
	return sc.DBDsn == "" && (sc.StoreInterval > 0 && sc.StoreFile != "")
}

func (sc ServerConfig) NeedCheckSign() bool {
	return sc.Key != ""
}

func GetConfigAgent() *AgentConfig {
	flag.StringVar(&address, "a", defaultAddress, "-a=<VALUE>")
	flag.DurationVar(&reportInterval, "r", defaultReportInterval, "-r=<VALUE>")
	flag.DurationVar(&pollInterval, "p", defaultPollInterval, "-p=<VALUE>")
	flag.StringVar(&key, "k", defaultKey, "-k=<KEY>")

	flag.Parse()

	return &AgentConfig{
		Address:        getEnvString("ADDRESS", address),
		ReportInterval: getEnvDuration("REPORT_INTERVAL", reportInterval),
		PollInterval:   getEnvDuration("POLL_INTERVAL", pollInterval),
		Key:            getEnvString("KEY", key),
	}
}

func (ac AgentConfig) NeedSign() bool {
	return ac.Key != ""
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
