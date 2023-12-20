package config

// Флаг -a отвечает за адрес запуска HTTP-сервера (значение может быть таким: localhost:8888).
// Флаг -b отвечает за базовый адрес результирующего сокращённого URL (значение: адрес сервера перед коротким URL, например http://localhost:8000/qsd54gFg).

import (
	"errors"
	"flag"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/caarlos0/env/v6"
)

const defaultRunAddr string = "localhost:8080"
const defaultBaseAddr string = "http://localhost"

type Options struct {
	runAddrOpt  string `env:"SERVER_ADDRESS"`
	baseAddrOpt string `env:"BASE_URL"`
}

var options Options

type NetAddress struct {
	host string
	port int
}

type Config struct {
	runAddr  NetAddress `env:"SERVER_ADDRESS"`
	baseAddr *url.URL   `env:"BASE_URL"`
}

func (c Config) GetRunAddr() string {
	return c.runAddr.host + ":" + strconv.Itoa(c.runAddr.port)
}

func (c Config) GetBaseAddr() string {
	return c.baseAddr.String()
}

func NewConfig() *Config {
	res := &Config{}

	res.ParseEnv()
	res.ParseFlags()

	hp := strings.Split(options.runAddrOpt, ":")
	if len(hp) != 2 {
		panic(errors.New("need address in a form host:port"))
	}
	port, err := strconv.Atoi(hp[1])
	if err != nil {
		panic(errors.New("port should be in numerical format"))
	}
	res.runAddr.host = hp[0]
	res.runAddr.port = port

	fmt.Println("0 RunAddr: " + res.GetRunAddr())

	res.baseAddr, err = url.Parse(options.baseAddrOpt)
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

	fmt.Println("0 BaseAddr: " + res.GetBaseAddr())
	return res
}

func (conf *Config) ParseFlags() error {
	flag.StringVar(&options.runAddrOpt, "a", defaultRunAddr, "address and port to run server")
	flag.StringVar(&options.baseAddrOpt, "b", defaultBaseAddr, "shortener address")
	flag.Parse()
	return nil
}

func (conf *Config) ParseEnv() error {
	err := env.Parse(conf)
	if err != nil {
		// OS environment parsing error
		return err
	}
	return nil
}
