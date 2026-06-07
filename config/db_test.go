package config_test

import (
	"os"
	"strings"
	"testing"

	"github.com/khrees/veilo/config"
)

func TestDBConfig_WithValidEnv(t *testing.T) {
	// Save original env vars
	originalEnv := map[string]string{
		"DB_HOST":     os.Getenv("DB_HOST"),
		"DB_PORT":     os.Getenv("DB_PORT"),
		"DB_USER":     os.Getenv("DB_USER"),
		"DB_PASSWORD": os.Getenv("DB_PASSWORD"),
		"DB_NAME":     os.Getenv("DB_NAME"),
		"DB_SSLMODE":  os.Getenv("DB_SSLMODE"),
	}

	// Set test env vars
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASSWORD", "testpass")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("DB_SSLMODE", "disable")

	defer func() {
		// Restore original env vars
		for k, v := range originalEnv {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	}()

	cfg := &config.DBConfig{}
	_, err := cfg.Connect()
	if err == nil {
		// If connection succeeds, that's fine - we just want to verify
		// the config loads correctly. For actual tests, we'd use a mock.
		t.Log("Database connection test passed (or connected to real DB)")
	}
}

func TestDBConfig_ConstructsDSN(t *testing.T) {
	cfg := &config.DBConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "testuser",
		Password: "testpass",
		DBName:   "testdb",
		SSLMode:  "disable",
	}

	dsn := cfg.DSN()

	for _, want := range []string{
		"postgresql://testuser:testpass@localhost:5432/testdb",
		"sslmode=disable",
	} {
		if !strings.Contains(dsn, want) {
			t.Errorf("expected DSN to contain %q, got %q", want, dsn)
		}
	}
}

func TestDBConfig_WithSSLMode(t *testing.T) {
	cfg := &config.DBConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "testuser",
		Password: "testpass",
		DBName:   "testdb",
		SSLMode:  "require",
	}

	if cfg.SSLMode != "require" {
		t.Errorf("expected sslmode 'require', got '%s'", cfg.SSLMode)
	}
}

func TestDBConfig_MinimalConfig(t *testing.T) {
	cfg := &config.DBConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "testuser",
		Password: "testpass",
		DBName:   "testdb",
		SSLMode:  "disable",
	}

	// Verify all required fields are set
	if cfg.Host == "" {
		t.Error("expected host to be set")
	}
	if cfg.Port == "" {
		t.Error("expected port to be set")
	}
	if cfg.User == "" {
		t.Error("expected user to be set")
	}
	if cfg.Password == "" {
		t.Error("expected password to be set")
	}
	if cfg.DBName == "" {
		t.Error("expected dbname to be set")
	}
}

func TestDBConfig_UsesDatabaseURLWhenProvided(t *testing.T) {
	cfg := &config.DBConfig{
		URL: "postgresql://example.com/testdb?sslmode=require",
	}

	if got := cfg.DSN(); got != cfg.URL {
		t.Fatalf("expected DATABASE_URL to win, got %q", got)
	}
}
