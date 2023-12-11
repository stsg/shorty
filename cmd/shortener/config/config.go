package config

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

type Cnf struct {
	// shortener run address
	RunAddr string `env:"SERVER_ADDRESS"`
	// Short URL server address
	Host string `env:"BASE_URL"`
}

var ShortyCnf Cnf

func ParseFlags() {
	flag.StringVar(&ShortyCnf.RunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&ShortyCnf.Host, "b", "http://localhost:8080/", "shortener address")

	flag.Parse()
}

func ParseEnv() {
	err := env.Parse(&ShortyCnf)
	if err != nil {
		panic(err)
	}
}
