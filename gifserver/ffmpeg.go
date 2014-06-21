package gifserver

import "os/exec"

func HasFFMPEG() bool {
	cmd := exec.Command("ffmpeg", "-h")
	err := cmd.Run()

	if err != nil {
		return false
	}

	return true
}
