package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/leafo/gifserver/gifserver"
)

var (
	configFname string
)

func init() {
	flag.StringVar(&configFname, "config",
		gifserver.DefaultConfigFname, "Path to json config")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: gifserver [OPTIONS]\n\nOptions:\n")
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()
	config := gifserver.LoadConfig(configFname)

	if !gifserver.HasFFMPEG() {
		log.Fatal("Could not find command `ffmpeg`, check to make sure it's install and available in $PATH")
	}

	gifserver.StartServer(config)
}
