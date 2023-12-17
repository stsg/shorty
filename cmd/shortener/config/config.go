package config

import (
	"flag"
	"fmt"
	"net"
	"net/url"
	"strings"

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

	fmt.Println("Flg RunAddr", ShortyCnf.RunAddr)
	fmt.Println("Flg Host", ShortyCnf.Host)

	return nil
}

func ParseEnv() error {
	err := env.Parse(&ShortyCnf)
	if err != nil {
		return err
	}
	fmt.Println("Env RunAddr", ShortyCnf.RunAddr)
	fmt.Println("Env Host", ShortyCnf.Host)
	return nil
}

func InitConfig() error {

	ParseFlags()
	ParseEnv()

	if !strings.HasPrefix(strings.ToLower(ShortyCnf.RunAddr), `http://`) {
		ShortyCnf.RunAddr = `http://` + ShortyCnf.RunAddr
	}

	runAddressURL, err := url.Parse(ShortyCnf.RunAddr)
	if err != nil {
		return err
	}

	RunAddress = ""
	host, port, _ := net.SplitHostPort(runAddressURL.Host)
	if host == "" {
		RunAddress += "localhost"
	} else {
		RunAddress += host
	}
	if port == "" {
		RunAddress += ":8080"
	} else {
		RunAddress += ":" + port
	}

	hostURL, err := url.Parse(ShortyCnf.Host)
	if err != nil {
		return err
	}
	if !strings.HasPrefix(strings.ToLower(ShortyCnf.Host), `http://`) {
		ShortyCnf.RunAddr = `http://` + ShortyCnf.Host
	}
	Host = ""
	host, port, _ = net.SplitHostPort(hostURL.Host)
	if host == "" {
		Host += "localhost"
	} else {
		Host += host
	}
	if port == "" {
		Host += ":8080"
	} else {
		Host += ":" + port
	}

	return nil
}

func InitConfig() error {

	ParseFlags()

	if !strings.HasPrefix(strings.ToLower(ShortyCnf.RunAddr), `http://`) {
		ShortyCnf.RunAddr = `http://` + ShortyCnf.RunAddr
	}

	runAddressURL, err := url.Parse(ShortyCnf.RunAddr)
	if err != nil {
		return err
	}

	RunAddress = ""
	host, port, _ := net.SplitHostPort(runAddressURL.Host)
	if host == "" {
		RunAddress += "localhost"
	} else {
		RunAddress += host
	}
	if port == "" {
		RunAddress += ":8080"
	} else {
		RunAddress += ":" + port
	}

	hostURL, err := url.Parse(ShortyCnf.Host)
	if err != nil {
		return err
	}
	if !strings.HasPrefix(strings.ToLower(ShortyCnf.Host), `http://`) {
		ShortyCnf.RunAddr = `http://` + ShortyCnf.Host
	}
	Host = ""
	host, port, _ = net.SplitHostPort(hostURL.Host)
	if host == "" {
		Host += "localhost"
	} else {
		Host += host
	}
	if port == "" {
		Host += ":8080"
	} else {
		Host += ":" + port
	}

	return nil
}
