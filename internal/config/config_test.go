package config

import (
	"flag"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetConfigServer(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want *ServerConfig
	}{
		{
			name: "Create server config from env variables test",
			env: map[string]string{
				"ADDRESS":        "127.0.0.1:8000",
				"STORE_INTERVAL": "250s",
				"STORE_FILE":     "/tmp/metrics-db.json",
				"RESTORE":        "0",
				"KEY":            "qwerty",
				"DATABASE_DSN":   "postgresql://user:pass@host/db",
			},
			want: &ServerConfig{
				Address:       "127.0.0.1:8000",
				StoreInterval: 250 * time.Second,
				StoreFile:     "/tmp/metrics-db.json",
				Restore:       false,
				Key:           "qwerty",
				DBDsn:         "postgresql://user:pass@host/db",
			},
		},
		{
			name: "Create server config from default values test",
			env:  map[string]string{},
			want: &ServerConfig{
				Address:       "127.0.0.1:8080",
				StoreInterval: 300 * time.Second,
				StoreFile:     "/tmp/devops-metrics-db.json",
				Restore:       true,
				Key:           "",
				DBDsn:         "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.env {
				err := os.Setenv(k, v)
				require.Nil(t, err)
			}
			defer func() {
				for k := range tt.env {
					err := os.Unsetenv(k)
					require.Nil(t, err)
				}
			}()

			assert.Equal(t, tt.want, GetConfigServer())

			flag.CommandLine = flag.NewFlagSet("", flag.ExitOnError)
			flag.CommandLine.Init("", flag.ContinueOnError)
		})
	}
}

func TestGetConfigAgent(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want *AgentConfig
	}{
		{
			name: "Create agent config from env variables test",
			env: map[string]string{
				"ADDRESS":         "127.0.0.1:8000",
				"REPORT_INTERVAL": "25s",
				"POLL_INTERVAL":   "10s",
				"KEY":             "qwerty",
				"RATE_LIMIT":      "5",
			},
			want: &AgentConfig{
				Address:        "127.0.0.1:8000",
				ReportInterval: 25 * time.Second,
				PollInterval:   10 * time.Second,
				Key:            "qwerty",
				RateLimit:      5,
			},
		},
		{
			name: "Create agent config from default values test",
			env:  map[string]string{},
			want: &AgentConfig{
				Address:        "127.0.0.1:8080",
				ReportInterval: 10 * time.Second,
				PollInterval:   2 * time.Second,
				Key:            "",
				RateLimit:      4,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.env {
				err := os.Setenv(k, v)
				require.Nil(t, err)
			}
			defer func() {
				for k := range tt.env {
					err := os.Unsetenv(k)
					require.Nil(t, err)
				}
			}()

			assert.Equal(t, tt.want, GetConfigAgent())

			flag.CommandLine = flag.NewFlagSet("", flag.ExitOnError)
			flag.CommandLine.Init("", flag.ContinueOnError)
		})
	}
}
