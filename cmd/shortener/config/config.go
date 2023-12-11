package config

import (
	"flag"
	"net"
	"net/url"
	"strings"
)

type Cnf struct {
	// shortener run address
	RunAddr string
	// Short URL server address
	Host string
}

var ShortyCnf Cnf
var RunAddress string
var Host string

func ParseFlags() {
	flag.StringVar(&ShortyCnf.RunAddr, "a", "http://localhost:8080", "address and port to run server")
	flag.StringVar(&ShortyCnf.Host, "b", "http://localhost/", "shortener address")

	flag.Parse()
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
