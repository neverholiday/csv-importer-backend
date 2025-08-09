package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/kelseyhightower/envconfig"
	"github.com/stretchr/testify/assert"
)

func TestEnvCfg_EnvironmentVariables(t *testing.T) {
	// Test with valid environment variables
	os.Setenv("CSV_IMPORTER_DB_HOST", "localhost")
	os.Setenv("CSV_IMPORTER_DB_PORT", "5432")
	os.Setenv("CSV_IMPORTER_DB_USER", "testuser")
	os.Setenv("CSV_IMPORTER_DB_PASSWORD", "testpass")
	os.Setenv("CSV_IMPORTER_DB_NAME", "testdb")
	defer func() {
		os.Unsetenv("CSV_IMPORTER_DB_HOST")
		os.Unsetenv("CSV_IMPORTER_DB_PORT")
		os.Unsetenv("CSV_IMPORTER_DB_USER")
		os.Unsetenv("CSV_IMPORTER_DB_PASSWORD")
		os.Unsetenv("CSV_IMPORTER_DB_NAME")
	}()

	var cfg EnvCfg
	err := envconfig.Process("CSV_IMPORTER", &cfg)
	assert.NoError(t, err)
	assert.Equal(t, "localhost", cfg.DBHost)
	assert.Equal(t, 5432, cfg.DBPort)
	assert.Equal(t, "testuser", cfg.DBUser)
	assert.Equal(t, "testpass", cfg.DBPassword)
	assert.Equal(t, "testdb", cfg.DBName)
}

func TestEnvCfg_MissingRequiredVariables(t *testing.T) {
	// Clear all environment variables
	vars := []string{
		"CSV_IMPORTER_DB_HOST",
		"CSV_IMPORTER_DB_PORT",
		"CSV_IMPORTER_DB_USER",
		"CSV_IMPORTER_DB_PASSWORD",
		"CSV_IMPORTER_DB_NAME",
	}
	
	for _, v := range vars {
		os.Unsetenv(v)
	}

	var cfg EnvCfg
	err := envconfig.Process("CSV_IMPORTER", &cfg)
	assert.Error(t, err, "Should fail when required environment variables are missing")
}

func TestEnvCfg_PartiallyMissingVariables(t *testing.T) {
	// Set some but not all required variables
	os.Setenv("CSV_IMPORTER_DB_HOST", "localhost")
	os.Setenv("CSV_IMPORTER_DB_PORT", "5432")
	// Missing USER, PASSWORD, NAME
	defer func() {
		os.Unsetenv("CSV_IMPORTER_DB_HOST")
		os.Unsetenv("CSV_IMPORTER_DB_PORT")
	}()

	var cfg EnvCfg
	err := envconfig.Process("CSV_IMPORTER", &cfg)
	assert.Error(t, err, "Should fail when some required environment variables are missing")
}

func TestEnvCfg_InvalidPortValue(t *testing.T) {
	os.Setenv("CSV_IMPORTER_DB_HOST", "localhost")
	os.Setenv("CSV_IMPORTER_DB_PORT", "invalid_port")
	os.Setenv("CSV_IMPORTER_DB_USER", "testuser")
	os.Setenv("CSV_IMPORTER_DB_PASSWORD", "testpass")
	os.Setenv("CSV_IMPORTER_DB_NAME", "testdb")
	defer func() {
		os.Unsetenv("CSV_IMPORTER_DB_HOST")
		os.Unsetenv("CSV_IMPORTER_DB_PORT")
		os.Unsetenv("CSV_IMPORTER_DB_USER")
		os.Unsetenv("CSV_IMPORTER_DB_PASSWORD")
		os.Unsetenv("CSV_IMPORTER_DB_NAME")
	}()

	var cfg EnvCfg
	err := envconfig.Process("CSV_IMPORTER", &cfg)
	assert.Error(t, err, "Should fail when port is not a valid integer")
}

func TestEnvCfg_EmptyValues(t *testing.T) {
	// Test with empty string values
	os.Setenv("CSV_IMPORTER_DB_HOST", "")
	os.Setenv("CSV_IMPORTER_DB_PORT", "5432")
	os.Setenv("CSV_IMPORTER_DB_USER", "")
	os.Setenv("CSV_IMPORTER_DB_PASSWORD", "")
	os.Setenv("CSV_IMPORTER_DB_NAME", "")
	defer func() {
		os.Unsetenv("CSV_IMPORTER_DB_HOST")
		os.Unsetenv("CSV_IMPORTER_DB_PORT")
		os.Unsetenv("CSV_IMPORTER_DB_USER")
		os.Unsetenv("CSV_IMPORTER_DB_PASSWORD")
		os.Unsetenv("CSV_IMPORTER_DB_NAME")
	}()

	var cfg EnvCfg
	err := envconfig.Process("CSV_IMPORTER", &cfg)
	// envconfig typically treats empty strings as missing for required fields
	assert.Error(t, err, "Should fail when required fields are empty")
}

func TestEnvCfg_DefaultPrefix(t *testing.T) {
	// Test that the correct prefix is used
	os.Setenv("WRONG_PREFIX_DB_HOST", "localhost")
	os.Setenv("CSV_IMPORTER_DB_HOST", "correct_host")
	os.Setenv("CSV_IMPORTER_DB_PORT", "5432")
	os.Setenv("CSV_IMPORTER_DB_USER", "testuser")
	os.Setenv("CSV_IMPORTER_DB_PASSWORD", "testpass")
	os.Setenv("CSV_IMPORTER_DB_NAME", "testdb")
	defer func() {
		os.Unsetenv("WRONG_PREFIX_DB_HOST")
		os.Unsetenv("CSV_IMPORTER_DB_HOST")
		os.Unsetenv("CSV_IMPORTER_DB_PORT")
		os.Unsetenv("CSV_IMPORTER_DB_USER")
		os.Unsetenv("CSV_IMPORTER_DB_PASSWORD")
		os.Unsetenv("CSV_IMPORTER_DB_NAME")
	}()

	var cfg EnvCfg
	err := envconfig.Process("CSV_IMPORTER", &cfg)
	assert.NoError(t, err)
	assert.Equal(t, "correct_host", cfg.DBHost, "Should use CSV_IMPORTER prefix")
}

func TestEnvCfg_NumericPortValues(t *testing.T) {
	testCases := []struct {
		name        string
		portValue   string
		expectError bool
		expectedPort int
	}{
		{
			name:         "Valid standard port",
			portValue:    "5432",
			expectError:  false,
			expectedPort: 5432,
		},
		{
			name:         "Valid alternative port",
			portValue:    "3306",
			expectError:  false,
			expectedPort: 3306,
		},
		{
			name:        "Zero port",
			portValue:   "0",
			expectError: false,
			expectedPort: 0,
		},
		{
			name:        "Negative port",
			portValue:   "-1",
			expectError: false,
			expectedPort: -1,
		},
		{
			name:        "Very large port",
			portValue:   "65535",
			expectError: false,
			expectedPort: 65535,
		},
		{
			name:        "Port too large",
			portValue:   "65536",
			expectError: false,
			expectedPort: 65536, // envconfig will parse this, app logic should validate
		},
		{
			name:        "Float value",
			portValue:   "5432.5",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			os.Setenv("CSV_IMPORTER_DB_HOST", "localhost")
			os.Setenv("CSV_IMPORTER_DB_PORT", tc.portValue)
			os.Setenv("CSV_IMPORTER_DB_USER", "testuser")
			os.Setenv("CSV_IMPORTER_DB_PASSWORD", "testpass")
			os.Setenv("CSV_IMPORTER_DB_NAME", "testdb")
			defer func() {
				os.Unsetenv("CSV_IMPORTER_DB_HOST")
				os.Unsetenv("CSV_IMPORTER_DB_PORT")
				os.Unsetenv("CSV_IMPORTER_DB_USER")
				os.Unsetenv("CSV_IMPORTER_DB_PASSWORD")
				os.Unsetenv("CSV_IMPORTER_DB_NAME")
			}()

			var cfg EnvCfg
			err := envconfig.Process("CSV_IMPORTER", &cfg)
			
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedPort, cfg.DBPort)
			}
		})
	}
}

// Test the database connection string formatting
func TestDatabaseConnectionString(t *testing.T) {
	cfg := EnvCfg{
		DBHost:     "localhost",
		DBPort:     5432,
		DBUser:     "testuser",
		DBPassword: "testpass",
		DBName:     "testdb",
	}

	expected := "host=localhost port=5432 user=testuser password=testpass dbname=testdb sslmode=disable TimeZone=UTC"
	actual := formatConnectionString(cfg)
	assert.Equal(t, expected, actual)
}

func TestDatabaseConnectionString_SpecialCharacters(t *testing.T) {
	cfg := EnvCfg{
		DBHost:     "db.example.com",
		DBPort:     3306,
		DBUser:     "user@domain",
		DBPassword: "pass word!@#$%",
		DBName:     "my-database",
	}

	expected := "host=db.example.com port=3306 user=user@domain password=pass word!@#$% dbname=my-database sslmode=disable TimeZone=UTC"
	actual := formatConnectionString(cfg)
	assert.Equal(t, expected, actual)
}

func TestDatabaseConnectionString_EmptyValues(t *testing.T) {
	cfg := EnvCfg{
		DBHost:     "",
		DBPort:     0,
		DBUser:     "",
		DBPassword: "",
		DBName:     "",
	}

	expected := "host= port=0 user= password= dbname= sslmode=disable TimeZone=UTC"
	actual := formatConnectionString(cfg)
	assert.Equal(t, expected, actual)
}

// Helper function to format connection string (extracted from main for testing)
func formatConnectionString(cfg EnvCfg) string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable TimeZone=UTC",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
	)
}