package ffmpeg

import (
	"bytes"
	"fmt"
	"os/exec"
)

func Transcode720p(inPath string, fileName string) error {

	type Variant struct {
		Name    string
		Scale   string
		Bitrate string
		CRF     int
		Preset  string
	}

	var outputs []Variant = []Variant{
		{Name: "1080p", Scale: "scale=-2:1080", CRF: 23, Preset: "fast"},
		{Name: "720p", Scale: "scale=-2:720", CRF: 23, Preset: "fast"},
		{Name: "490p", Scale: "scale=-2:480", CRF: 23, Preset: "fast"},
	}

	for _, o := range outputs {
		outPath := fmt.Sprintf("./processed/%s/%s", fileName, o.Name)
		cmd := exec.Command("ffmpeg",
			"-i", inPath,
			"-map", "0:v:0",
			"-map", "0:a:0",
			"-vf", o.Scale,
			"-c:v", "libx264",
			"-crf", fmt.Sprint(o.CRF),
			"-preset", o.Preset,
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

	}

	return nil

}
