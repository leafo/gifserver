package gifserver

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"
)

func TestLimitedReader(t *testing.T) {
	arr := [1024]byte{}

	reader := bytes.NewReader(arr[:])
	_, err := ioutil.ReadAll(limitedReader(reader, 512))
	if err == nil {
		t.Error("should fail on limited reader")
	}

	reader = bytes.NewReader(arr[:])
	_, err = ioutil.ReadAll(limitedReader(reader, 1025))

	if err != nil {
		t.Error("should not fail on limited reader")
	}
}

func TestTryWriter(t *testing.T) {
	fakeWriter := func(p []byte) (int, error) {
		return 0, fmt.Errorf("always fails")
	}

	arr := [1024]byte{}

	_, err := bytes.NewReader(arr[:]).WriteTo(writerClosure(fakeWriter))
	if err == nil {
		t.Error("fake writer should alway fail")
	}

	safeWriter := tryWriter(writerClosure(fakeWriter))

	_, err = bytes.NewReader(arr[:]).WriteTo(safeWriter)
	if err != nil {
		t.Error("should not fail on try writer")
	}
}
