package gifserver

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
)

const testGifFname = "../test/hello.gif"

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

func TestExtract(t *testing.T) {
	input, _ := os.Open(testGifFname)
	defer input.Close()

	dir, err := prepareConversion(input)

	if err != nil {
		t.Fatal(err.Error())
	}

	defer cleanDir(dir)

	err = extractGif(dir)

	if err != nil {
		t.Error(err.Error())
	}

}

func TestConvertGifMP4(t *testing.T) {
	if !HasFFMPEG() {
		t.Skip("missing ffmpeg, can't continue")
	}

	input, _ := os.Open(testGifFname)
	dir, _ := prepareConversion(input)

	defer cleanDir(dir)

	_, err := convertGifToMP4(dir)

	if err != nil {
		t.Error(err.Error())
	}
}

func TestConvertFramesMP4(t *testing.T) {
	if !HasFFMPEG() {
		t.Skip("missing ffmpeg, can't continue")
	}

	input, _ := os.Open(testGifFname)
	dir, _ := prepareConversion(input)

	defer cleanDir(dir)
	extractGif(dir)

	_, err := convertFramesToMP4(dir)

	if err != nil {
		t.Error(err.Error())
	}
}

func TestConvertFramesOGV(t *testing.T) {
	if !HasFFMPEG() {
		t.Skip("missing ffmpeg, can't continue")
	}

	input, _ := os.Open(testGifFname)
	dir, _ := prepareConversion(input)

	defer cleanDir(dir)
	extractGif(dir)

	_, err := convertFramesToOGV(dir)

	if err != nil {
		t.Error(err.Error())
	}
}
