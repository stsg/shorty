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
	RunAddrOpt  string `env:"SERVER_ADDRESS"`
	BaseAddrOpt string `env:"BASE_URL"`
	FileStorOpt string `env:"FILE_STORAGE_PATH"`
}
type NetAddress struct {
	host string
	port int
}

type Config struct {
	runAddr  NetAddress
	baseAddr *url.URL
	fileStor string
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

	flag.StringVar(&opt.RunAddrOpt, "a", defaultRunAddr, "address and port to run server")
	flag.StringVar(&opt.BaseAddrOpt, "b", defaultBaseAddr, "shortener address")
	flag.StringVar(&opt.FileStorOpt, "f", defaultFileStor, "file storage path")
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

	if opt.FileStorOpt != "" {
		res.fileStor = opt.FileStorOpt
	} else {
		res.fileStor = "/dev/null"
	}
	return res
}
