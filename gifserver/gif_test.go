package gifserver

import (
	"bytes"
	"io/ioutil"
	"testing"
)

const testGifFname = "../test/hello.gif"

func TestOpen(t *testing.T) {
	_, err := loadGif(testGifFname)
	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestCheckDimensions(t *testing.T) {
	gifBytes, _ := ioutil.ReadFile(testGifFname)

	err := checkDimensions(bytes.NewReader(gifBytes), 0, 0)
	if err != nil {
		t.Error("0,0 dimensions should not have failure")
	}

	err = checkDimensions(bytes.NewReader(gifBytes), 10, 10)
	if err == nil {
		t.Error("should fail check dimensions")
	}

	err = checkDimensions(bytes.NewReader(gifBytes), 0, 10)
	if err == nil {
		t.Error("should fail check dimensions")
	}


	err = checkDimensions(bytes.NewReader(gifBytes), 300, 300)
	if err != nil {
		t.Error("should not fail check dimensions: " + err.Error())
	}
}

func TestExtraction(t *testing.T) {
	gif, _ := loadGif(testGifFname)
	dir, err := extractGif(gif)
	defer cleanDir(dir)

	if err != nil {
		t.Error(err.Error())
	}
}

func TestConvertMP4(t *testing.T) {
	if !HasFFMPEG() {
		t.Skip("missing ffmpeg, can't continue")
	}

	gif, _ := loadGif(testGifFname)
	dir, _ := extractGif(gif)
	defer cleanDir(dir)

	_, err := convertToMP4(dir)

	if err != nil {
		t.Error(err.Error())
	}
}

func TestConvertOGV(t *testing.T) {
	if !HasFFMPEG() {
		t.Skip("missing ffmpeg, can't continue")
	}

	gif, _ := loadGif(testGifFname)
	dir, _ := extractGif(gif)
	defer cleanDir(dir)

	_, err := convertToOGV(dir)

	if err != nil {
		t.Error(err.Error())
	}
}


