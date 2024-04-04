// Package config - config package for shorty service, URL shortener application
package config

import (
	"errors"
	"flag"
	"net/url"
	"strconv"
	"strings"

	"github.com/caarlos0/env/v6"
)

const defaultRunAddr string = "localhost:8080"
const defaultBaseAddr string = "http://localhost:8080"
const defaultFileStorage string = "/tmp/short-url-db.json"

// should be in form "host=localhost port=5432 user=postgres dbname=postgres password=postgres sslmode=disable"
const defaultDBStorage string = ""

// Options class definition defines a struct holds Options
// with four fields: RunAddrOpt, BaseAddrOpt, FileStorageOpt, and DBStorageOpt.
// Each field is tagged with an env tag,
// which specifies the name of the environment variable
// that should be used to set the value of that field.
type Options struct {
	RunAddrOpt     string `env:"SERVER_ADDRESS"`
	BaseAddrOpt    string `env:"BASE_URL"`
	FileStorageOpt string `env:"FILE_STORAGE_PATH"`
	DBStorageOpt   string `env:"DATABASE_DSN"`
}

var opt Options

// NetAddress is a struct that holds URL host and port
type NetAddress struct {
	host string
	port int
}

// Config is a struct that holds Application configuration
type Config struct {
	baseAddr    *url.URL
	storageType string
	fileStorage string
	dbStorage   string
	runAddr     NetAddress
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
//
// If any of the flags are missing or have invalid values, the function panics.
//
// The function also parses the following environment variables using the "github.com/caarlos0/env" package:
// - "RUN_ADDR": the address and port to run the server.
// - "BASE_ADDR": the shortener address.
// - "FILE_STORAGE": the file storage path.
// - "DB_STORAGE": the database DSN.
//
// If any of the environment variables are missing or have invalid values, the function panics.
//
// The function returns the created Config object.
func NewConfig() Config {
	res := Config{}
	// opt := Options{}

	// flag.StringVar(&res.opt.RunAddrOpt, "a", defaultRunAddr, "address and port to run server")
	// flag.StringVar(&res.opt.BaseAddrOpt, "b", defaultBaseAddr, "shortener address")
	// flag.StringVar(&res.opt.FileStorageOpt, "f", defaultFileStorage, "file storage path")
	// flag.StringVar(&res.opt.DBStorageOpt, "d", defaultDBStorage, "database DSN")
	flag.Parse()

	err := env.Parse(&opt)
	if err != nil {
		// OS environment parsing error
		panic(errors.New("cannot parse OS environment"))
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

	return res
}
func init() {
	flag.StringVar(&opt.RunAddrOpt, "a", defaultRunAddr, "address and port to run server")
	flag.StringVar(&opt.BaseAddrOpt, "b", defaultBaseAddr, "shortener address")
	flag.StringVar(&opt.FileStorageOpt, "f", defaultFileStorage, "file storage path")
	flag.StringVar(&opt.DBStorageOpt, "d", defaultDBStorage, "database DSN")
	// fmt.Println("config.init()")
}
