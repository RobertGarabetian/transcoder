package ffmpeg

import (
	"bytes"
	"fmt"
	"os/exec"
)

func Transcode720p(inPath string, outPath string) error {
	cmd := exec.Command("ffmpeg",
		"-i", inPath,
		"-map", "0:v:0",
		"-map", "0:a:0",
		"-vf", "scale=trunc(iw/2)*2:trunc(ih/2)*2",
		"-c:v", "libx264",
		"-crf", "23",
		"-preset", "fast",
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
