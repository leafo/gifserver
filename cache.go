package main

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

func (self *FileCache) Put(fname string, reader io.Reader) (int64, error) {
	writeCloser, err := self.PutWriter(fname)

	if err != nil {
		return 0, err
	}

	defer writeCloser.Close()
	return io.Copy(writeCloser, reader)
}

func (self *FileCache) PutWriter(fname string) (io.WriteCloser, error) {
	target := path.Join(self.Dir, fname)
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

func (self *FileCache) Get(fname string) (io.ReadCloser, error) {
	return os.Open(path.Join(self.Dir, fname))
}
