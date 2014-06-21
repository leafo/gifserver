package main

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
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

const maxBytes = 1024 * 1024 * 5 // 5mb

type basicHandler func(w http.ResponseWriter, r *http.Request) error

func (fn basicHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := fn(w, r); err != nil {
		log.Print("ERROR: ", err.Error())
		http.Error(w, err.Error(), 500)
	}
}

type readerClosure func(p []byte) (int, error)
type writerClosure func(p []byte) (int, error)

func (fn readerClosure) Read(p []byte) (int, error) {
	return fn(p)
}

func (fn writerClosure) Write(p []byte) (int, error) {
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

func tryWriter(writer io.Writer) writerClosure {
	failed := false
	return func(p []byte) (int, error) {
		if failed {
			return len(p), nil
		}

		written, err := writer.Write(p)

		if err != nil {
			log.Print("Try writer failed, continuing...")
			failed = true
			written = len(p)
			err = nil
		}

		return written, err
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

func checkSignature(r *http.Request) error {
	params := r.URL.Query()
	sig := params.Get("sig")

	if sig == "" {
		return fmt.Errorf("Missing signature")
	}

	patt := regexp.MustCompile(`[?&]sig=[^?&]+`)

	toCheck := r.URL.Path
	strippedQuery := patt.ReplaceAllString(r.URL.RawQuery, "")

	if strippedQuery != "" {
		toCheck = toCheck + "?" + strippedQuery
	}

	mac := hmac.New(sha1.New, []byte("secret"))
	mac.Write([]byte(toCheck))
	expectedSig := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	if expectedSig != sig {
		return fmt.Errorf("Invalid signature")
	}

	return nil
}

func startServer(listenTo string) {
	http.Handle("/transcode", basicHandler(transcodeHandler))
	log.Print("Listening on ", listenTo)
	http.ListenAndServe(listenTo, nil)
}
