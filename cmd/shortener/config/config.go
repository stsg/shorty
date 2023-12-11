package config

import (
	"flag"
)

type Cnf struct {
	// shortener run address
	RunAddr string
	// Short URL server address
	Host string
}

var ShortyCnf Cnf

func ParseFlags() {
	flag.StringVar(&ShortyCnf.RunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&ShortyCnf.Host, "b", "http://localhost/", "shortener address")

	flag.Parse()
}
