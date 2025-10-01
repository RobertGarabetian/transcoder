package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"transcoder/internal/ffmpeg"
	"transcoder/internal/storage"
)

// CORS middleware so frontend (localhost:3000) can call backend (localhost:8080)
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	// 1. Initialize S3 client (MinIO)
	s3Client, err := storage.NewS3Client(context.Background(), "transcoder.project")
	if err != nil {
		log.Fatal("failed to init S3 client:", err)
	}

	// 2. Define upload handler
	uploadHandler := func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseMultipartForm(10 << 20) // 10 MB max memory
		if err != nil {
			log.Println("parse error:", err)
			http.Error(w, "failed to parse form", http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("video")
		if err != nil {
			log.Println("formfile error:", err)
			http.Error(w, "failed to read file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Save temp copy locally
		tmpPath := "./" + header.Filename
		dst, err := os.Create(tmpPath)
		if err != nil {
			log.Println("os.Create error:", err)
			http.Error(w, "failed to create temp file", http.StatusInternalServerError)
			return
		}
		_, err = io.Copy(dst, file)
		dst.Close()
		if err != nil {
			log.Println("copy error:", err)
			http.Error(w, "failed to save temp file", http.StatusInternalServerError)
			return
		}

		ctx := context.Background()

		// Upload raw file to S3
		rawKey := "raw/" + header.Filename
		err = s3Client.UploadFile(ctx, rawKey, tmpPath)
		if err != nil {
			log.Println("s3 upload error:", err)
			http.Error(w, "failed to upload raw file", http.StatusInternalServerError)
			return
		}

		// Create processed output dir (local)
		processedDir := "./processed/" + header.Filename
		err = os.MkdirAll(processedDir, 0755)
		if err != nil {
			log.Println("mkdir error:", err)
			http.Error(w, "failed to create processed dir", http.StatusInternalServerError)
			return
		}

		// Run FFmpeg HLS transcode
		err = ffmpeg.TranscodeHLS(tmpPath, processedDir)
		if err != nil {
			log.Println("ffmpeg error:", err)
			http.Error(w, "failed to transcode video", http.StatusInternalServerError)
			return
		}

		// Upload all generated files (master.m3u8, variants, segments) to S3
		err = filepath.Walk(processedDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				relPath := filepath.Base(path)
				processedKey := fmt.Sprintf("processed/%s/%s", header.Filename, relPath)
				err = s3Client.UploadFile(ctx, processedKey, path)
				if err != nil {
					return fmt.Errorf("failed to upload %s: %w", path, err)
				}
			}
			return nil
		})
		if err != nil {
			log.Println("s3 upload processed error:", err)
			http.Error(w, "failed to upload processed files", http.StatusInternalServerError)
			return
		}

		// 🔥 Rewrite master.m3u8 to include signed URLs for .ts files
		masterPath := filepath.Join(processedDir, "master.m3u8")
		content, err := os.ReadFile(masterPath)
		if err != nil {
			log.Println("read master.m3u8 error:", err)
			http.Error(w, "failed to read master playlist", http.StatusInternalServerError)
			return
		}

		lines := strings.Split(string(content), "\n")
		for i, line := range lines {
			if strings.HasSuffix(line, ".ts") {
				tsKey := fmt.Sprintf("processed/%s/%s", header.Filename, line)
				signed, err := s3Client.GenerateSignedURL(ctx, tsKey, 15*time.Minute)
				if err != nil {
					log.Println("signed ts url error:", err)
					http.Error(w, "failed to sign segment url", http.StatusInternalServerError)
					return
				}
				lines[i] = signed
			}
		}

		newContent := strings.Join(lines, "\n")
		signedMasterPath := filepath.Join(processedDir, "master_signed.m3u8")
		os.WriteFile(signedMasterPath, []byte(newContent), 0644)

		// Upload rewritten playlist to S3
		signedMasterKey := fmt.Sprintf("processed/%s/master_signed.m3u8", header.Filename)
		err = s3Client.UploadFile(ctx, signedMasterKey, signedMasterPath)
		if err != nil {
			log.Println("upload signed master error:", err)
			http.Error(w, "failed to upload signed playlist", http.StatusInternalServerError)
			return
		}

		// Generate signed URL for the new playlist
		signedURL, err := s3Client.GenerateSignedURL(ctx, signedMasterKey, 15*time.Minute)
		if err != nil {
			log.Println("signed url error:", err)
			http.Error(w, "failed to sign playlist url", http.StatusInternalServerError)
			return
		}

		// ✅ Respond back to frontend
		resp := map[string]string{
			"raw":    rawKey,
			"master": signedURL,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}

	videoHandler := func(w http.ResponseWriter, r *http.Request) {
		url, err := s3Client.GenerateSignedURL(context.Background(), "processed/720pzzzz.MOV", 15*time.Minute)
		if err != nil {
			http.Error(w, "failed to generate signed URL", http.StatusInternalServerError)
			return
		}
		resp := map[string]string{
			"url": url}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		fmt.Println(w, url)
	}
	//3. Wire routes
	mux := http.NewServeMux()
	mux.HandleFunc("/upload", uploadHandler)
	mux.HandleFunc("/video", videoHandler)

	// 4. Start server
	log.Println("Server running at http://localhost:8080")
	if err := http.ListenAndServe(":8080", withCORS(mux)); err != nil {
		log.Fatal(err)
	}
}
