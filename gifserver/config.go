package gifserver

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

const DefaultConfigFname = "gifserver.json"

type config struct {
	Address        string
	Secret         string
	CacheDir       string
	MaxBytes       int
	MaxWidth       int
	MaxHeight      int
	MaxConcurrency int
}

var defaultConfig = config{
	Address:        ":9090",
	Secret:         "",
	CacheDir:       "gifcache",
	MaxBytes:       1024 * 1024 * 5, // 5mb
	MaxWidth:       512,
	MaxHeight:      512,
	MaxConcurrency: 0,
}

func LoadConfig(fname string) *config {
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
