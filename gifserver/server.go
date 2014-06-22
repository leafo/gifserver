package gifserver

import (
	"bytes"
	"crypto/md5"

	"encoding/hex"
	"fmt"
	"image/gif"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"sync"
	"time"
)

var (
	serverConfig     *config
	concurrencyLimit chan bool
)

const defaultFormat = "mp4"

var shared struct {
	sync.Mutex
	processing map[string]bool
}

var cache *FileCache

func init() {
	shared.processing = make(map[string]bool)
}

func keyBusy(key string) bool {
	shared.Lock()
	defer shared.Unlock()
	return shared.processing[key]
}

func lockKey(key string) {
	shared.Lock()
	defer shared.Unlock()
	shared.processing[key] = true
}

func unlockKey(key string) {
	shared.Lock()
	defer shared.Unlock()
	shared.processing[key] = false
}

type basicHandler func(w http.ResponseWriter, r *http.Request) error

func (fn basicHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := fn(w, r); err != nil {
		log.Print("ERROR: ", err.Error())
		http.Error(w, err.Error(), 500)
	}
}

func requestKey(url, format string) string {
	raw := md5.Sum([]byte(url))
	return hex.EncodeToString(raw[:]) + "." + format
}

func getConverter(format string) (string, converter, error) {
	if format == "" {
		format = defaultFormat
	}

	var c converter

	switch format {
	case "mp4":
		c = convertToMP4
	case "ogv":
		c = convertToOGV
	case "png":
		c = convertToFrame
	default:
		return "", nil, fmt.Errorf("Invalid format")
	}

	return format, c, nil
}

func getContentType(format string) string {
	switch format {
	case "mp4":
		return "video/mp4"
	case "ogv":
		return "video/ogg"
	case "png":
		return "image/png"
	}
	return ""
}

func transcodeHandler(w http.ResponseWriter, r *http.Request) error {
	params := r.URL.Query()
	err := checkSignature(r, serverConfig.Secret)
	if err != nil {
		return err
	}

	url := params.Get("url")

	if url == "" {
		return fmt.Errorf("Missing param: url")
	}

	urlPattern := regexp.MustCompile(`^\w+://`)
	if !urlPattern.MatchString(url) {
		url = "http://" + url
	}

	format, conv, err := getConverter(params.Get("format"))

	if err != nil {
		return err
	}

	key := requestKey(url, format)

	for keyBusy(key) {
		time.Sleep(time.Second / 5)
	}

	lockKey(key)
	defer unlockKey(key)

	// see if cache has it
	readCloser, err := cache.Get(key)

	if err == nil {
		log.Print("Hit cache for ", key)
		w.Header().Add("Content-type", getContentType(format))
		io.Copy(w, readCloser)
		defer readCloser.Close()
		return nil
	}

	res, err := http.Get(url)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	resBody := io.Reader(res.Body)

	// Check if content length is too large
	if serverConfig.MaxBytes > 0 {
		contentLengthStr := res.Header.Get("content-length")
		if contentLengthStr != "" {
			contentLength, err := strconv.Atoi(contentLengthStr)
			if err != nil && contentLength > serverConfig.MaxBytes {
				return fmt.Errorf("Image is too large (%d > %d)",
					contentLength, serverConfig.MaxBytes)
			}
		}

		resBody = limitedReader(resBody, serverConfig.MaxBytes)
	}

	gifData, err := ioutil.ReadAll(resBody)

	if err != nil {
		return err
	}

	err = checkDimensions(bytes.NewReader(gifData),
		serverConfig.MaxWidth, serverConfig.MaxHeight)

	if err != nil {
		return err
	}

	if serverConfig.MaxConcurrency > 0 {
		<-concurrencyLimit
		defer func() {
			concurrencyLimit <- true
		}()
	}

	gif, err := gif.DecodeAll(bytes.NewReader(gifData))

	if err != nil {
		return err
	}

	dir, err := extractGif(gif)
	defer cleanDir(dir)

	if err != nil {
		return err
	}

	vidFname, err := conv(dir)

	if err != nil {
		return err
	}

	file, err := os.Open(vidFname)

	if err != nil {
		return err
	}

	cacheWriter, err := cache.PutWriter(key)

	if err != nil {
		return err
	}

	w.Header().Add("Content-type", getContentType(format))

	multi := io.MultiWriter(tryWriter(w), cacheWriter)
	bytes, err := io.Copy(multi, file)
	log.Print("Wrote ", bytes, " bytes")

	return err
}

func StartServer(_config *config) {
	serverConfig = _config
	cache = NewFileCache(serverConfig.CacheDir)

	if serverConfig.MaxConcurrency > 0 {
		concurrencyLimit = make(chan bool, serverConfig.MaxConcurrency)
		for i := 0; i < serverConfig.MaxConcurrency; i++ {
			concurrencyLimit <- true
		}
	}

	http.Handle("/transcode", basicHandler(transcodeHandler))
	log.Print("Listening on ", serverConfig.Address)
	http.ListenAndServe(serverConfig.Address, nil)
}
