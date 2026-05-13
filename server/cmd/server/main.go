package main

import (
	"context"
	"log"
	"net/http"
	"time"
	"transcoder/internal/handlers"
	"transcoder/internal/storage"
)

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
	s3Client, err := storage.NewS3Client(context.Background(), "transcoder.project")
	if err != nil {
		log.Fatal("failed to init S3 client:", err)
	}

	go func() {
		ticker := time.NewTicker(2 * time.Hour)
		defer ticker.Stop()

		log.Println("S3 auto-cleanup goroutine started. Will delete all files every 2 hours.")

		for range ticker.C {
			ctx := context.Background()
			log.Println("Starting scheduled S3 bucket cleanup...")
			if err := s3Client.DeleteAllObjects(ctx); err != nil {
				log.Printf("Error during scheduled cleanup: %v", err)
			} else {
				log.Println("Scheduled cleanup completed successfully")
			}
		}
	}()

	h := &handlers.Handler{S3: s3Client}

	mux := http.NewServeMux()
	mux.HandleFunc("/upload", h.Upload)
	mux.HandleFunc("/list", h.List)
	mux.HandleFunc("/cleanup", h.Cleanup)

	log.Printf("Server running at :8080")
	if err := http.ListenAndServe(":8080", withCORS(mux)); err != nil {
		log.Fatal(err)
	}
}
