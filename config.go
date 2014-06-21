package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

var defaultConfigFname = "gifserver.json"

type config struct {
	Secret   string
	MaxBytes int
}

var defaultConfig = config{
	Secret:   "",
	MaxBytes: 1024 * 1024 * 5, // 5mb
}

func loadConfig(fname string) *config {
	c := defaultConfig
	if fname == "" {
		return &c
	}

	jsonBlob, err := ioutil.ReadFile(fname)
	if err == nil {
		err = json.Unmarshal(jsonBlob, &c)

		if err != nil {
			log.Fatal("Failed parsing config: ", fname, ": ", err.Error())
		}
	} else {
		log.Print(err.Error())
	}

	return &c
}
