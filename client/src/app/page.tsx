"use client";

import { useState, useRef, useEffect } from "react";
import Hls from "hls.js";

interface TranscodingProgress {
  format: string;
  status: "pending" | "processing" | "completed" | "error";
  progress: number;
  error?: string;
}

export default function Home() {
  const [file, setFile] = useState<File | null>(null);
  const [isUploading, setIsUploading] = useState(false);
  const [uploadProgress, setUploadProgress] = useState(0);
  const [overallStatus, setOverallStatus] = useState("");
  const [videoUrl, setVideoUrl] = useState<string | null>(null);
  const videoRef = useRef<HTMLVideoElement | null>(null);

  const [transcodingProgress, setTranscodingProgress] = useState<
    TranscodingProgress[]
  >([
    { format: "1080p", status: "pending", progress: 0 },
    { format: "720p", status: "pending", progress: 0 },
    { format: "490p", status: "pending", progress: 0 },
  ]);

  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return "0 Bytes";
    const k = 1024;
    const sizes = ["Bytes", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const selectedFile = e.target.files?.[0] || null;
    setFile(selectedFile);
    if (selectedFile) {
      setOverallStatus("");
      setTranscodingProgress([
        { format: "1080p", status: "pending", progress: 0 },
        { format: "720p", status: "pending", progress: 0 },
        { format: "490p", status: "pending", progress: 0 },
      ]);
    }
  };

  const handleUpload = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!file) {
      setOverallStatus("Please choose a file first.");
      return;
    }

    setIsUploading(true);
    setUploadProgress(0);
    setOverallStatus("Uploading file...");

    const formData = new FormData();
    formData.append("video", file);

    try {
      // simulate progress bar locally
      const progressInterval = setInterval(() => {
        setUploadProgress((prev) => {
          if (prev >= 90) {
            clearInterval(progressInterval);
            return prev;
          }
          return prev + Math.random() * 10;
        });
      }, 200);

      const res = await fetch("http://localhost:8080/upload", {
        method: "POST",
        body: formData,
      });
      const data = await res.json();
      setVideoUrl(data.master);
      clearInterval(progressInterval);
      setUploadProgress(100);

      if (res.ok) {
        setOverallStatus("✅ Upload successful! Processing video formats...");
        // fake transcoding progress simulation (optional)
        const formats = ["1080p", "720p", "490p"];
        formats.forEach((format, index) => {
          setTimeout(() => {
            setTranscodingProgress((prev) =>
              prev.map((p) =>
                p.format === format
                  ? { ...p, status: "processing" as const }
                  : p
              )
            );

            const interval = setInterval(() => {
              setTranscodingProgress((prev) =>
                prev.map((p) => {
                  if (p.format === format && p.status === "processing") {
                    const newProgress = Math.min(
                      p.progress + Math.random() * 15,
                      100
                    );
                    if (newProgress >= 100) {
                      clearInterval(interval);
                      return { ...p, progress: 100, status: "completed" };
                    }
                    return { ...p, progress: newProgress };
                  }
                  return p;
                })
              );
            }, 300);
          }, index * 1000);
        });

        setTimeout(() => {
          setOverallStatus("✅ All formats processed successfully!");
        }, 5000);
      } else {
        setOverallStatus("❌ Upload failed: " + res.statusText);
        setIsUploading(false);
      }
    } catch (err) {
      setOverallStatus("❌ Error: " + (err as Error).message);
      setIsUploading(false);
    }
  };

  const resetForm = () => {
    setFile(null);
    setIsUploading(false);
    setUploadProgress(0);
    setOverallStatus("");
    setVideoUrl(null);
    setTranscodingProgress([
      { format: "1080p", status: "pending", progress: 0 },
      { format: "720p", status: "pending", progress: 0 },
      { format: "490p", status: "pending", progress: 0 },
    ]);
    if (fileInputRef.current) {
      fileInputRef.current.value = "";
    }
  };

  const fileInputRef = useRef<HTMLInputElement>(null);

  // attach hls.js player when we have a signed URL
  useEffect(() => {
    if (videoUrl && videoRef.current) {
      if (Hls.isSupported()) {
        const hls = new Hls();
        hls.loadSource(videoUrl);
        hls.attachMedia(videoRef.current);
      } else if (
        videoRef.current.canPlayType("application/vnd.apple.mpegurl")
      ) {
        // Safari supports HLS natively
        videoRef.current.src = videoUrl;
      }
    }
  }, [videoUrl]);

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 dark:from-gray-900 dark:to-gray-800">
      <div className="container mx-auto px-4 py-8">
        <div className="max-w-4xl mx-auto">
          <div className="text-center mb-8">
            <h1 className="text-4xl font-bold text-gray-900 dark:text-white mb-2">
              Video Transcoder
            </h1>
            <p className="text-gray-600 dark:text-gray-300">
              Upload your video and get multiple optimized formats
            </p>
          </div>

          <div className="bg-white dark:bg-gray-800 rounded-xl shadow-lg p-8">
            <form onSubmit={handleUpload} className="space-y-6">
              <div className="space-y-4">
                <input
                  ref={fileInputRef}
                  type="file"
                  accept="video/*"
                  onChange={handleFileChange}
                  disabled={isUploading}
                />
                {file && (
                  <div>
                    <p>{file.name}</p>
                    <p>
                      {formatFileSize(file.size)} • {file.type}
                    </p>
                  </div>
                )}
              </div>

              <button type="submit" disabled={!file || isUploading}>
                {isUploading ? "Processing..." : "Start Transcoding"}
              </button>
            </form>

            {uploadProgress > 0 && (
              <div>Upload Progress: {Math.round(uploadProgress)}%</div>
            )}

            {overallStatus && <p>{overallStatus}</p>}

            {/* Video Player */}
            {videoUrl && (
              <div className="mt-6">
                <video ref={videoRef} controls className="w-full rounded-lg" />
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
