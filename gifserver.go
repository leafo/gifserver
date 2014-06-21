package main

import (
	"fmt"
	"image/gif"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
)

type gifData struct {
	width  int
	height int
	frames int
}

func loadGif(fname string) (*gif.GIF, error) {
	file, err := os.Open(fname)

	if err != nil {
		return nil, err
	}

	defer file.Close()

	return gif.DecodeAll(file)
}

func extractGif(gif *gif.GIF) (string, error) {
	dir, err := ioutil.TempDir("", "gifserver")

	if err != nil {
		return "", err
	}

	log.Print("Extracting ", len(gif.Image), "frames")

	for i, image := range gif.Image {
		dest := path.Join(dir, fmt.Sprintf("frame_%05d.png", i))
		fmt.Println("Extracting frame", i, "to", dest)
		file, err := os.Create(dest)
		if err != nil {
			return "", err
		}

		defer file.Close()

		png.Encode(file, image)
	}

	return dir, nil
}

func cleanDir(dir string) error {
	log.Print("Removing", dir)
	return os.Remove(dir)
}

// log1 "Making ${out_base}.mp4..."
// ffmpeg -i "$pattern" -pix_fmt yuv420p -vf 'scale=trunc(in_w/2)*2:trunc(in_h/2)*2' "${out_base}.mp4"

func convertToMP4(dir string) (string, error) {
	outFname := "out.mp4"
	pattern := "frame_%05d.png"
	cmd := exec.Command("ffmpeg",
		"-i", pattern,
		"-pix_fmt", "yuv420p",
		"-vf", "scale=trunc(in_w/2)*2:trunc(in_h/2)*2'",
		outFname)

	cmd.Dir = dir
	err := cmd.Run()

	if err != nil {
		return "", err
	}

	return path.Join(dir, outFname), nil
}

// log1 "Making ${out_base}.ogv..."
// ffmpeg -i "$pattern" -q 5 -pix_fmt yuv420p "${out_base}.ogv"

func convertToOGV(dir string) (string, error) {
	outFname := "out.ogv"

	pattern := "frame_%05d.png"
	cmd := exec.Command("ffmpeg",
		"-i", pattern,
		"-q", "5",
		"-pix_fmt", "yuv420p",
		outFname)

	cmd.Dir = dir
	err := cmd.Run()

	if err != nil {
		return "", err
	}

	return path.Join(dir, outFname), nil
}

func copyFile(src, dest string) error {
	log.Print("Copying ", src, " to ", dest)

	input, err := os.Open(src)
	if err != nil {
		return err
	}

	output, err := os.Create(dest)

	if err != nil {
		return err
	}

	defer output.Close()

	_, err = io.Copy(output, input)

	if err != nil {
		return err
	}

	defer input.Close()
	return nil
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Missing file")
	}

	gif, err := loadGif(os.Args[1])

	if err != nil {
		log.Fatal(err.Error())
	}

	dir, err := extractGif(gif)
	if err != nil {
		log.Fatal(err.Error())
	}

	defer cleanDir(dir)
	vid, err := convertToOGV(dir)
	if err != nil {
		log.Fatal(err.Error())
	}

	fmt.Println(vid)
	if len(os.Args) > 2 {
		copyFile(vid, os.Args[2])
	}
}
