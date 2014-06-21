package gifserver

import (
	"crypto/md5"

	"encoding/hex"
	"fmt"
	"image/gif"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"sync"
	"time"
)

var serverConfig *config

var shared struct {
	sync.Mutex
	processing map[string]bool
}

var cache *FileCache

func init() {
	shared.processing = make(map[string]bool)
	cache = NewFileCache("gifcache")
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

func transcodeHandler(w http.ResponseWriter, r *http.Request) error {
	params := r.URL.Query()
	url := params.Get("url")

	if url == "" {
		return fmt.Errorf("Missing param: url")
	}

	urlPattern := regexp.MustCompile(`^\w+://`)
	if !urlPattern.MatchString(url) {
		url = "http://" + url
	}

	key := requestKey(url, "mp4")

	for keyBusy(key) {
		time.Sleep(time.Second / 5)
	}

	lockKey(key)
	defer unlockKey(key)

	// see if cache has it
	readCloser, err := cache.Get(key)

	if err == nil {
		log.Print("Hit cache for ", key)
		w.Header().Add("Content-type", "video/mp4")
		io.Copy(w, readCloser)
		defer readCloser.Close()
		return nil
	}

	res, err := http.Get(url)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	// Check if content length is too large
	contentLengthStr := res.Header.Get("content-length")
	if contentLengthStr != "" {
		contentLength, err := strconv.Atoi(contentLengthStr)
		if err != nil && contentLength > serverConfig.MaxBytes {
			return fmt.Errorf("Image is too large (%d > %d)", contentLength, serverConfig.MaxBytes)
		}
	}

	gif, err := gif.DecodeAll(limitedReader(res.Body, serverConfig.MaxBytes))

	if err != nil {
		return err
	}

	dir, err := extractGif(gif)
	defer cleanDir(dir)

	if err != nil {
		return err
	}

	vidFname, err := convertToMP4(dir)

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

	w.Header().Add("Content-type", "video/mp4")

	multi := io.MultiWriter(tryWriter(w), cacheWriter)
	bytes, err := io.Copy(multi, file)
	log.Print("Wrote ", bytes, " bytes")

	return err
}

func StartServer(listenTo string, _config *config) {
	serverConfig = _config
	http.Handle("/transcode", basicHandler(transcodeHandler))
	log.Print("Listening on ", listenTo)
	http.ListenAndServe(listenTo, nil)
}
