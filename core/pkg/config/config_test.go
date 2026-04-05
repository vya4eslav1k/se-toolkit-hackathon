package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad_DefaultValues(t *testing.T) {
	// Clear environment variables
	os.Unsetenv("DB_HOST")
	os.Unsetenv("DB_PORT")
	os.Unsetenv("DB_USER")
	os.Unsetenv("DB_PASSWORD")
	os.Unsetenv("DB_NAME")
	os.Unsetenv("SERVER_PORT")
	os.Unsetenv("COOKIE_SECRET")

	cfg, err := Load()

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "localhost", cfg.DBHost)
	assert.Equal(t, 5432, cfg.DBPort)
	assert.Equal(t, "postgres", cfg.DBUser)
	assert.Equal(t, "postgres", cfg.DBPassword)
	assert.Equal(t, "polls", cfg.DBName)
	assert.Equal(t, 8080, cfg.ServerPort)
	assert.Equal(t, "super-secret-key-change-in-production", cfg.CookieSecret)
}

func TestLoad_EnvironmentVariables(t *testing.T) {
	os.Setenv("DB_HOST", "test-db-host")
	os.Setenv("DB_PORT", "5433")
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASSWORD", "testpass")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("COOKIE_SECRET", "test-secret")

	cfg, err := Load()

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "test-db-host", cfg.DBHost)
	assert.Equal(t, 5433, cfg.DBPort)
	assert.Equal(t, "testuser", cfg.DBUser)
	assert.Equal(t, "testpass", cfg.DBPassword)
	assert.Equal(t, "testdb", cfg.DBName)
	assert.Equal(t, 9090, cfg.ServerPort)
	assert.Equal(t, "test-secret", cfg.CookieSecret)

	// Clean up
	os.Unsetenv("DB_HOST")
	os.Unsetenv("DB_PORT")
	os.Unsetenv("DB_USER")
	os.Unsetenv("DB_PASSWORD")
	os.Unsetenv("DB_NAME")
	os.Unsetenv("SERVER_PORT")
	os.Unsetenv("COOKIE_SECRET")
}

func TestLoad_InvalidDBPort(t *testing.T) {
	os.Setenv("DB_PORT", "invalid")

	cfg, err := Load()

	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "invalid DB_PORT")

	os.Unsetenv("DB_PORT")
}

func TestLoad_InvalidServerPort(t *testing.T) {
	os.Setenv("SERVER_PORT", "invalid")

	cfg, err := Load()

	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "invalid SERVER_PORT")

	os.Unsetenv("SERVER_PORT")
}

func TestConfig_DBConnString(t *testing.T) {
	cfg := &Config{
		DBHost:     "localhost",
		DBPort:     5432,
		DBUser:     "postgres",
		DBPassword: "password",
		DBName:     "testdb",
	}

	connStr := cfg.DBConnString()
	expected := "host=localhost port=5432 user=postgres password=password dbname=testdb sslmode=disable"

	assert.Equal(t, expected, connStr)
}
