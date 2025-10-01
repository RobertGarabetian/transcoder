package ffmpeg

import (
	"bytes"
	"fmt"
	"os/exec"
)

func TranscodeHLS(inPath string, outDir string) error {
	cmd := exec.Command("ffmpeg",
		"-i", inPath,

		// force 8-bit
		"-pix_fmt", "yuv420p",

		// only take first video and first audio stream
		"-map", "0:v:0",
		"-map", "0:a:0",

		// 1080p
		"-filter:v:0", "scale=-2:1080",
		"-c:v:0", "libx264", "-profile:v:0", "high", "-crf", "20", "-preset", "veryfast",
		"-b:v:0", "5000k", "-maxrate:v:0", "5350k", "-bufsize:v:0", "7500k",

		// 720p
		"-filter:v:1", "scale=-2:720",
		"-c:v:1", "libx264", "-profile:v:1", "main", "-crf", "20", "-preset", "veryfast",
		"-b:v:1", "2800k", "-maxrate:v:1", "2996k", "-bufsize:v:1", "4200k",

		// 480p
		"-filter:v:2", "scale=-2:480",
		"-c:v:2", "libx264", "-profile:v:2", "baseline", "-crf", "20", "-preset", "veryfast",
		"-b:v:2", "800k", "-maxrate:v:2", "856k", "-bufsize:v:2", "1200k",

		// audio
		"-c:a", "aac", "-ar", "48000", "-b:a", "128k",

		"-f", "hls",
		"-hls_time", "6",
		"-hls_playlist_type", "vod",
		"-hls_segment_filename", fmt.Sprintf("%s/stream_%%v/data%%03d.ts", outDir),
		"-master_pl_name", "master.m3u8",
		"-var_stream_map", "v:0,a:0 v:1,a:0 v:2,a:0",

		fmt.Sprintf("%s/stream_%%v.m3u8", outDir),
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		fmt.Println("ffmpeg failed:", stderr.String())
		return err
	}

	return nil
}
