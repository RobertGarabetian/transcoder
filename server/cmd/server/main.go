package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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
	s3Client, err := storage.NewS3Client(
		"http://localhost:9000", // MinIO API endpoint
		"us-east-1",             // fake region, MinIO just needs *something*
		"admin",                 // access key
		"admin123",              // secret key
		"videos",                // bucket name you created in MinIO
	)
	if err != nil {
		log.Fatal("failed to init S3 client:", err)
	}

	// 2. Define upload handler
	uploadHandler := func(w http.ResponseWriter, r *http.Request) {
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
		key := "raw/" + header.Filename
		err = s3Client.UploadFile(ctx, key, tmpPath)
		if err != nil {
			log.Println("s3 upload error:", err) // 👈 important
			http.Error(w, "failed to upload to bucket", http.StatusInternalServerError)
			return
		}

		processedPath := "./processed/" + header.Filename
		err = ffmpeg.Transcode720p(tmpPath, header.Filename)
		if err != nil {
			log.Println("ffmpeg error:", err)
			http.Error(w, "failed to transcode video", http.StatusInternalServerError)
		}

		processedKey := "processed/" + header.Filename
		err = s3Client.UploadFile(ctx, processedKey, processedPath)
		if err != nil {
			log.Println("s3 upload processed error: ", err)
			http.Error(w, "Error uploading transcoded file to s3", http.StatusInternalServerError)
			return
		}

		_ = os.Remove(tmpPath)
		_ = os.Remove(processedPath)

		fmt.Fprintf(w, "✅ Raw file: raw/%s\n✅ Transcoded file: %s\n", header.Filename, processedKey)
	}

	// 3. Wire routes
	mux := http.NewServeMux()
	mux.HandleFunc("/upload", uploadHandler)

	// 4. Start server
	log.Println("Server running at http://localhost:8080")
	if err := http.ListenAndServe(":8080", withCORS(mux)); err != nil {
		log.Fatal(err)
	}
}
