package main

import (
	"fmt"
	"io"
)

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
			failed = true
			written = len(p)
			err = nil
		}

		return written, err
	}
}
