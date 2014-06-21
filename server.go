package main

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
)

var shared struct {
	sync.Mutex
	processing map[string]bool
}

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

const maxBytes = 1024 * 1024 * 5 // 5mb

type basicHandler func(w http.ResponseWriter, r *http.Request) error

func (fn basicHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := fn(w, r); err != nil {
		http.Error(w, err.Error(), 500)
	}
}

type readerClosure func(p []byte) (int, error)

func (fn readerClosure) Read(p []byte) (int, error) {
	return fn(p)
}

func limitedReader(reader io.Reader, maxBytes int) readerClosure {
	remainingBytes := maxBytes

	return func(p []byte) (int, error) {
		bytesRead, err := reader.Read(p)

		remainingBytes -= bytesRead
		if remainingBytes < 0 {
			return 0, fmt.Errorf("Image is too large (> %d)", maxBytes)
		}

		return bytesRead, err
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
	if keyBusy(key) {
		return fmt.Errorf("Key is busy")
	}

	lockKey(key)
	defer unlockKey(key)

	res, err := http.Get(url)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	// Check if content length is too large
	contentLengthStr := res.Header.Get("content-length")
	if contentLengthStr != "" {
		contentLength, err := strconv.Atoi(contentLengthStr)
		if err != nil && contentLength > maxBytes {
			return fmt.Errorf("Image is too large (%d > %d)", contentLength, maxBytes)
		}
	}

	gif, err := gif.DecodeAll(limitedReader(res.Body, maxBytes))

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

	w.Header().Add("Content-type", "video/mp4")
	io.Copy(w, file)

	return nil
}

func startServer(listenTo string) {
	http.Handle("/transcode", basicHandler(transcodeHandler))
	log.Print("Listening on ", listenTo)
	http.ListenAndServe(listenTo, nil)
}
