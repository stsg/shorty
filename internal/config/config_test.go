package config

import (
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Returns a Config object with default values when no flags or environment variables are set.
func TestNewConfig_DefaultValues(t *testing.T) {
	config := &Config{
		runAddr:     NetAddress{"localhost", 8080},
		baseAddr:    &url.URL{Scheme: "http", Host: "localhost:8080"},
		storageType: "",
		fileStorage: "/tmp/short-url-db.json",
		// dbStorage:   "host=localhost port=5432 user=postgres dbname=postgres password=postgres sslmode=disable",
		dbStorage: "",
	}
	// Invoke the NewConfig function
	// config := NewConfig()

	// Assert that the run address is set to the default value
	assert.Equal(t, defaultRunAddr, config.GetRunAddr())

	// Assert that the base address is set to the default value
	assert.Equal(t, defaultBaseAddr, config.GetBaseAddr())

	// Assert that the storage type is empty
	assert.Equal(t, "", config.GetStorageType())

	// Assert that the file storage path is set to the default value
	assert.Equal(t, defaultFileStorage, config.GetFileStorage())

	// Assert that the DB storage is empty
	assert.Equal(t, "", config.GetDBStorage())
}

// Returns a Config object with the specified values when all flags and environment variables are set.
func TestNewConfig_SpecifiedValues(t *testing.T) {
	config := &Config{
		runAddr:     NetAddress{"localhost", 8080},
		baseAddr:    &url.URL{Scheme: "http", Host: "localhost:8080"},
		storageType: "file",
		fileStorage: "/tmp/short-url-db.json",
		dbStorage:   "/dev/null",
	}
	assert.Equal(t, *config, NewConfig())
}

// Panics when OS environment parsing fails.
func TestNewConfig_EnvironmentParsingError1(t *testing.T) {
	// Set an invalid environment variable
	os.Setenv("SERVER_ADDRESS", "invalid")

	// Assert that the NewConfig function panics with an error message
	assert.PanicsWithError(t, "need address in a form host:port",
		func() {
			NewConfig()
		},
	)
}

func TestNewConfig_EnvironmentParsingError2(t *testing.T) {
	// Set an invalid environment variable
	os.Setenv("SERVER_ADDRESS", "invalid:XXX")

	// Assert that the NewConfig function panics with an error message
	assert.PanicsWithError(t, "port should be in numerical format",
		func() {
			NewConfig()
		},
	)
}
