package handlers

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

type Handler struct {
	S3 *storage.S3Client
}

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20)
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

	filename := filepath.Base(header.Filename)

	tmpPath := "./" + filename
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

	rawKey := "raw/" + filename
	err = h.S3.UploadFile(ctx, rawKey, tmpPath)
	if err != nil {
		log.Println("s3 upload error:", err)
		http.Error(w, "failed to upload raw file", http.StatusInternalServerError)
		return
	}

	processedDir := "./processed/" + filename
	err = os.MkdirAll(processedDir, 0755)
	if err != nil {
		log.Println("mkdir error:", err)
		http.Error(w, "failed to create processed dir", http.StatusInternalServerError)
		return
	}

	err = ffmpeg.TranscodeHLS(tmpPath, processedDir)
	if err != nil {
		log.Println("ffmpeg error:", err)
		http.Error(w, "failed to transcode video", http.StatusInternalServerError)
		return
	}

	err = filepath.Walk(processedDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			relPath := filepath.Base(path)
			processedKey := fmt.Sprintf("processed/%s/%s", filename, relPath)
			err = h.S3.UploadFile(ctx, processedKey, path)
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
			tsKey := fmt.Sprintf("processed/%s/%s", filename, line)
			signed, err := h.S3.GenerateSignedURL(ctx, tsKey, 15*time.Minute)
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

	signedMasterKey := fmt.Sprintf("processed/%s/master_signed.m3u8", filename)
	err = h.S3.UploadFile(ctx, signedMasterKey, signedMasterPath)
	if err != nil {
		log.Println("upload signed master error:", err)
		http.Error(w, "failed to upload signed playlist", http.StatusInternalServerError)
		return
	}

	signedURL, err := h.S3.GenerateSignedURL(ctx, signedMasterKey, 15*time.Minute)
	if err != nil {
		log.Println("signed url error:", err)
		http.Error(w, "failed to sign playlist url", http.StatusInternalServerError)
		return
	}

	resp := map[string]string{
		"raw":    rawKey,
		"master": signedURL,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

	go func() {
		os.Remove(tmpPath)
		os.RemoveAll(processedDir)
	}()
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := context.Background()
	keys, err := h.S3.ListObjects(ctx)
	if err != nil {
		log.Printf("Error listing objects: %v", err)
		http.Error(w, "failed to list objects", http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"count":   len(keys),
		"objects": keys,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) Cleanup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := context.Background()
	log.Println("Manual cleanup triggered via /cleanup endpoint")

	keysBefore, _ := h.S3.ListObjects(ctx)

	err := h.S3.DeleteAllObjects(ctx)
	if err != nil {
		log.Printf("Error during manual cleanup: %v", err)
		http.Error(w, fmt.Sprintf("cleanup failed: %v", err), http.StatusInternalServerError)
		return
	}

	keysAfter, _ := h.S3.ListObjects(ctx)

	resp := map[string]interface{}{
		"status":    "success",
		"deleted":   len(keysBefore),
		"remaining": len(keysAfter),
		"message":   fmt.Sprintf("Deleted %d objects", len(keysBefore)),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
