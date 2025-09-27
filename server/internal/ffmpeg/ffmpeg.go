package ffmpeg

import (
	"bytes"
	"fmt"
	"os/exec"
)

type Variant struct {
	Name    string
	Scale   string
	Bitrate string
	CRF     int
	Preset  string
}

func Transcode720p(inPath string, name string, preset string, crf int, scale string, outPath string) error {

	cmd := exec.Command("ffmpeg",
		"-i", inPath,
		"-map", "0:v:0",
		"-map", "0:a:0",
		"-vf", scale,
		"-c:v", "libx264",
		"-crf", fmt.Sprint(crf),
		"-preset", preset,
		"-c:a", "aac",
		outPath,
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		fmt.Println("ffmpeg failed:", stderr.String()) // print ffmpeg’s error logs
		return err
	}

	return nil

}
