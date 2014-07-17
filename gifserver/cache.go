package gifserver

import (
	"io"
	"log"
	"os"
	"path"
)

// filesystem cache

type FileCache struct {
	Dir string
}

func NewFileCache(dir string) *FileCache {
	return &FileCache{
		Dir: dir,
	}
}

func (cache *FileCache) Put(fname string, reader io.Reader) (int64, error) {
	writeCloser, err := cache.PutWriter(fname)

	if err != nil {
		return 0, err
	}

	defer writeCloser.Close()
	return io.Copy(writeCloser, reader)
}

func (cache *FileCache) PutWriter(fname string) (io.WriteCloser, error) {
	target := path.Join(cache.Dir, fname)
	log.Print("Writing ", target)

	err := os.MkdirAll(path.Dir(target), 0755)

	if err != nil {
		return nil, err
	}

	file, err := os.Create(target)

	if err != nil {
		return nil, err
	}

	return file, err
}

func (cache *FileCache) Get(fname string) (*os.File, error) {
	return os.Open(path.Join(cache.Dir, fname))
}
