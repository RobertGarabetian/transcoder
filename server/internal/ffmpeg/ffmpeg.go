package ffmpeg

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
)

// TranscodeHLS takes an input video and outputs an HLS playlist + .ts segments
// into the given output directory. Example:
//
//	./processed/zzzz/master.m3u8
//	./processed/zzzz/720p_001.ts
//	./processed/zzzz/720p_002.ts
func TranscodeHLS(inPath string, outDir string) error {
	// Ensure output paths
	playlist := filepath.Join(outDir, "master.m3u8")
	segments := filepath.Join(outDir, "720p_%03d.ts")

	cmd := exec.Command("ffmpeg",
		"-i", inPath,
		"-map", "0:v:0", "-map", "0:a:0",
		"-vf", "scale=-2:720,format=yuv420p",
		"-c:v", "libx264", "-crf", "20", "-preset", "veryfast",
		"-c:a", "aac", "-ar", "48000", "-b:a", "128k",
		"-f", "hls",
		"-hls_time", "6",
		"-hls_playlist_type", "vod",
		"-hls_segment_filename", segments,
		playlist,
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("ffmpeg failed: %v\nLogs: %s", err, stderr.String())
	}

	return nil
}
