package config

import (
	"flag"
	// "fmt"

	// "net"
	// "net/url"

	"github.com/caarlos0/env/v6"
)

type Cnf struct {
	// shortener run address
	RunAddr string `env:"SERVER_ADDRESS"`
	// Short URL server address
	Host string `env:"BASE_URL"`
}

var ShortyCnf Cnf
var RunAddress string
var Host string

func ParseFlags() error {
	flag.StringVar(&ShortyCnf.RunAddr, "a", "http://localhost:8080", "address and port to run server")
	flag.StringVar(&ShortyCnf.Host, "b", "http://localhost", "shortener address")

	flag.Parse()

	return nil
}

func ParseEnv() error {
	err := env.Parse(&ShortyCnf)
	if err != nil {
		return err
	}
	return nil
}

func InitConfig() error {

	ParseFlags()
	ParseEnv()

	return nil
}
