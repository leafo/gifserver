package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/leafo/gifserver/gifserver"
)

var (
	configFname string
)

func init() {
	flag.StringVar(&configFname, "config", gifserver.DefaultConfigFname, "Path to json config")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: gifserver [OPTIONS]\n\nOptions:\n")
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()
	config := gifserver.LoadConfig(configFname)
	gifserver.StartServer(":9090", config)
}
