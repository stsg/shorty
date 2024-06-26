// Package config - config package for shorty service, URL shortener application
package config

import (
	"encoding/json"
	"errors"
	"flag"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/caarlos0/env/v6"
	"github.com/stsg/shorty/internal/logger"
	"go.uber.org/zap"
)

const defaultRunAddr string = "localhost:8080"
const defaultBaseAddr string = "http://localhost:8080"
const defaultFileStorage string = "/tmp/short-url-db.json"

// should be in form "host=localhost port=5432 user=postgres dbname=postgres password=postgres sslmode=disable"
const defaultDBStorage string = ""
const defaultConfigFile string = ""

// Options class definition defines a struct holds Options
// with four fields: RunAddrOpt, BaseAddrOpt, FileStorageOpt, and DBStorageOpt.
// Each field is tagged with an env tag,
// which specifies the name of the environment variable
// that should be used to set the value of that field.
type Options struct {
	RunAddrOpt     string `env:"SERVER_ADDRESS" json:"server_address,omitempty"`
	BaseAddrOpt    string `env:"BASE_URL" json:"base_url,omitempty"`
	FileStorageOpt string `env:"FILE_STORAGE_PATH" json:"file_storage_path,omitempty"`
	DBStorageOpt   string `env:"DATABASE_DSN" json:"database_dsn,omitempty"`
	EnableHTTPS    bool   `env:"ENABLE_HTPPS" json:"enable_https,omitempty"`
	TrustedSubnet  string `env:"TRUSTED_SUBNET" json:"trusted_subnet,omitempty"`
	ConfigFile     string `env:"CONFIG"`
}

var opt Options

// NetAddress is a struct that holds URL host and port
type NetAddress struct {
	host string
	port int
}

// Config is a struct that holds Application configuration
type Config struct {
	baseAddr      *url.URL
	storageType   string
	fileStorage   string
	dbStorage     string
	runAddr       NetAddress
	enableHTTPS   bool
	trustedSubnet *net.IPNet
	configFile    string
}

// GetRunAddr returns the run address of the Config object.
//
// It concatenates the host and port values from the runAddr field
// of the Config struct and returns the resulting string.
//
// Returns:
// - string: The run address in the format "host:port".
func (conf Config) GetRunAddr() string {
	return conf.runAddr.host + ":" + strconv.Itoa(conf.runAddr.port)
}

// GetBaseAddr returns the base address as a string.
//
// No parameters.
// Returns a string.
func (conf Config) GetBaseAddr() string {
	return conf.baseAddr.String()
}

// GetStorageType returns the storage type from the Config.
//
// No parameters.
// Returns a string.
func (conf Config) GetStorageType() string {
	return conf.storageType
}

// GetFileStorage returns the file storage path from the Config struct.
//
// No parameters.
// Returns a string.
func (conf Config) GetFileStorage() string {
	return conf.fileStorage
}

// GetDBStorage returns the DB storage for the Config.
//
// No parameters.
// Returns a string.
func (conf Config) GetDBStorage() string {
	return conf.dbStorage
}

// GetEnableHTTPS returns the value of the enableHTTPS field from the Config struct.
//
// No parameters.
// Returns a boolean value.
func (conf Config) GetEnableHTTPS() bool {
	return conf.enableHTTPS
}

// GetTrustedSubnet returns the value of the trustedSubnet field from the Config struct.
//
// No parameters.
// Returns a boolean value.
func (conf Config) GetTrustedSubnet() *net.IPNet {
	return conf.trustedSubnet
}

// IsTrusted checks if the given IP address is trusted based on the trusted subnet in the Config struct.
//
// Parameters:
// - ip: a string representing the IP address to check.
//
// Returns:
// - a boolean indicating whether the IP address is trusted or not.
func (conf Config) IsTrusted(ip string) bool {
	if conf.trustedSubnet == nil {
		return true
	}
	ipAddr := net.ParseIP(ip)
	return conf.trustedSubnet.Contains(ipAddr)
}

// GetConfigFile returns the config file path from the Config struct.
//
// No parameters.
// Returns a string.
func (conf Config) GetConfigFile() string {
	return conf.configFile
}

// NewConfig creates a new Config object by parsing command line flags and environment variables.
//
// It returns a Config object with the following fields:
// - runAddr: a struct containing the host and port to run the server.
// - baseAddr: a URL object representing the shortener address.
// - fileStorage: the path to the file storage.
// - dbStorage: the DSN of the database.
// - storageType: the type of storage being used, either "file" or "db".
//
// The function parses the following command line flags:
// - "-a": the address and port to run the server.
// - "-b": the shortener address.
// - "-f": the file storage path.
// - "-d": the database DSN.
// - "-s": enable HTTPS.
//
// If any of the flags are missing or have invalid values, the function panics.
//
// The function also parses the following environment variables using the "github.com/caarlos0/env" package:
// - "RUN_ADDR": the address and port to run the server.
// - "BASE_ADDR": the shortener address.
// - "FILE_STORAGE": the file storage path.
// - "DB_STORAGE": the database DSN.
// - "ENABLE_HTTPS": enable HTTPS.
//
// If any of the environment variables are missing or have invalid values, the function panics.
//
// The function returns the created Config object.
func NewConfig() Config {
	logger := logger.Get()
	res := Config{}
	flag.Parse()

	err := env.Parse(&opt)
	if err != nil {
		// OS environment parsing error
		panic(errors.New("cannot parse OS environment"))
	}

	if opt.ConfigFile != "" {
		logger.Info("reading config file", zap.String("path", opt.ConfigFile))
		configData, err := os.ReadFile(opt.ConfigFile)
		if err != nil {
			panic(errors.New("cannot read config file"))
		}
		err = json.Unmarshal(configData, &opt)
		if err != nil {
			panic(errors.New("cannot parse config file"))
		}
		logger.Info("config file parsed")
	}

	hp := strings.Split(opt.RunAddrOpt, ":")
	if len(hp) != 2 {
		panic(errors.New("need address in a form host:port"))
	}
	port, err := strconv.Atoi(hp[1])
	if err != nil {
		panic(errors.New("port should be in numerical format"))
	}
	res.runAddr.host = hp[0]
	res.runAddr.port = port

	opt.BaseAddrOpt = strings.TrimSuffix(opt.BaseAddrOpt, "/")
	res.baseAddr, err = url.Parse(opt.BaseAddrOpt)
	if err != nil {
		panic(errors.New("cannot parse base address"))
	}
	if !res.baseAddr.IsAbs() {
		res.baseAddr.Scheme = "http"
		if opt.EnableHTTPS {
			res.baseAddr.Scheme = "https"
		}
		res.baseAddr.Host = "localhost"
		if res.baseAddr.Path[0] != '/' {
			res.baseAddr.Path = "/" + res.baseAddr.Path
		}
	}

	if opt.FileStorageOpt != "" {
		res.fileStorage = opt.FileStorageOpt
		res.storageType = "file"
	} else {
		res.fileStorage = "/dev/null"
	}

	if opt.DBStorageOpt != "" {
		res.dbStorage = opt.DBStorageOpt
		res.storageType = "db"
	} else {
		res.dbStorage = "/dev/null"
	}

	if opt.EnableHTTPS {
		res.enableHTTPS = true
	}

	if opt.TrustedSubnet != "" {
		_, res.trustedSubnet, err = net.ParseCIDR(opt.TrustedSubnet)
		if err != nil {
			panic(errors.New("cannot parse trusted subnet"))
		}
	}

	return res
}

// init initializes the configuration by setting up flag options.
//
// No parameters.
// No return types.
func init() {
	flag.StringVar(&opt.RunAddrOpt, "a", defaultRunAddr, "address and port to run server")
	flag.StringVar(&opt.BaseAddrOpt, "b", defaultBaseAddr, "shortener address")
	flag.StringVar(&opt.FileStorageOpt, "f", defaultFileStorage, "file storage path")
	flag.StringVar(&opt.DBStorageOpt, "d", defaultDBStorage, "database DSN")
	flag.BoolVar(&opt.EnableHTTPS, "s", false, "enable HTTPS")
	flag.StringVar(&opt.TrustedSubnet, "t", "", "trusted subnet")
	flag.StringVar(&opt.ConfigFile, "c", defaultConfigFile, "config file path")
}
