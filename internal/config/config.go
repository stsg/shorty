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
const defaultFileStor string = "/tmp/short-url-db.json"

type options struct {
	runAddrOpt  string `env:"SERVER_ADDRESS"`
	baseAddrOpt string `env:"BASE_URL"`
	fileStorOpt string `env:"FILE_STORAGE_PATH"`
}

type NetAddress struct {
	host string
	port int
}

type Config struct {
	runAddr  NetAddress // `env:"SERVER_ADDRESS"`
	baseAddr *url.URL   // `env:"BASE_URL"`
	fileStor string     // `env:"FILE_STORAGE_PATH"`
}

func (conf Config) GetRunAddr() string {
	return conf.runAddr.host + ":" + strconv.Itoa(conf.runAddr.port)
}

func (conf Config) GetBaseAddr() string {
	return conf.baseAddr.String()
}

func (conf Config) GetFileStor() string {
	return conf.fileStor
}

func NewConfig() Config {
	res := Config{}
	opt := options{}

	flag.StringVar(&opt.runAddrOpt, "a", defaultRunAddr, "address and port to run server")
	flag.StringVar(&opt.baseAddrOpt, "b", defaultBaseAddr, "shortener address")
	flag.StringVar(&opt.fileStorOpt, "f", defaultFileStor, "file strorager path")
	flag.Parse()
	// res.ParseFlags()

	err := env.Parse(&opt)
	if err != nil {
		// OS environment parsing error
		panic(errors.New("cannot parse OS environment"))
	}

	hp := strings.Split(opt.runAddrOpt, ":")
	if len(hp) != 2 {
		panic(errors.New("need address in a form host:port"))
	}
	port, err := strconv.Atoi(hp[1])
	if err != nil {
		panic(errors.New("port should be in numerical format"))
	}
	res.runAddr.host = hp[0]
	res.runAddr.port = port

	opt.baseAddrOpt = strings.TrimSuffix(opt.baseAddrOpt, "/")
	res.baseAddr, err = url.Parse(opt.baseAddrOpt)
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

	res.fileStor = opt.fileStorOpt

	return res
}
