"use client";

import { useState, useRef, useEffect } from "react";
import Hls from "hls.js";

export default function Home() {
  const [file, setFile] = useState<File | null>(null);
  const [isUploading, setIsUploading] = useState(false);
  const [uploadProgress, setUploadProgress] = useState(0);
  const [status, setStatus] = useState("");
  const [videoUrl, setVideoUrl] = useState<string | null>(null);

  const fileInputRef = useRef<HTMLInputElement>(null);
  const videoRef = useRef<HTMLVideoElement | null>(null);

  const API_URL = process.env.NEXT_PUBLIC_API_URL;

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const selectedFile = e.target.files?.[0] || null;
    setFile(selectedFile);
    setStatus("");
  };

  const handleUpload = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!file) {
      setStatus("Please choose a file first.");
      return;
    }

    setIsUploading(true);
    setUploadProgress(0);
    setStatus("Uploading file...");

    const formData = new FormData();
    formData.append("video", file);

    try {
      const res = await fetch(`${API_URL}/upload`, {
        method: "POST",
        body: formData,
      });

      const data = await res.json();

      if (!res.ok) throw new Error(data.error || res.statusText);

      setVideoUrl(data.master);
      setUploadProgress(100);
      setStatus("✅ Upload successful! Processing video...");
    } catch (err) {
      setStatus("❌ Error: " + (err as Error).message);
    } finally {
      setIsUploading(false);
    }
  };

  const resetForm = () => {
    setFile(null);
    setIsUploading(false);
    setUploadProgress(0);
    setStatus("");
    setVideoUrl(null);
    if (fileInputRef.current) fileInputRef.current.value = "";
  };

  useEffect(() => {
    if (videoUrl && videoRef.current) {
      if (Hls.isSupported()) {
        const hls = new Hls();
        hls.loadSource(videoUrl);
        hls.attachMedia(videoRef.current);
        return () => hls.destroy();
      } else if (
        videoRef.current.canPlayType("application/vnd.apple.mpegurl")
      ) {
        // Safari native HLS
        videoRef.current.src = videoUrl;
      }
    }
  }, [videoUrl]);

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 dark:from-gray-900 dark:to-gray-800">
      <div className="container mx-auto px-4 py-8">
        <div className="max-w-3xl mx-auto">
          <div className="text-center mb-8">
            <h1 className="text-4xl font-bold text-gray-900 dark:text-white mb-2">
              Video Transcoder
            </h1>
            <p className="text-gray-600 dark:text-gray-300">
              Upload your video and get an optimized streamable version
            </p>
          </div>

          <div className="bg-white dark:bg-gray-800 rounded-xl shadow-lg p-8">
            <form onSubmit={handleUpload} className="space-y-6">
              <div>
                <input
                  ref={fileInputRef}
                  type="file"
                  accept="video/*"
                  onChange={handleFileChange}
                  disabled={isUploading}
                />
                {file && (
                  <p className="mt-2 text-sm text-gray-600 dark:text-gray-400">
                    {file.name} • {(file.size / 1024 / 1024).toFixed(2)} MB
                  </p>
                )}
              </div>

              <button
                type="submit"
                disabled={!file || isUploading}
                className="px-4 py-2 bg-indigo-600 text-white rounded-lg disabled:opacity-50"
              >
                {isUploading ? "Uploading..." : "Start Upload"}
              </button>

              {file && !isUploading && (
                <button
                  type="button"
                  onClick={resetForm}
                  className="ml-2 px-4 py-2 bg-gray-300 dark:bg-gray-700 text-gray-900 dark:text-white rounded-lg"
                >
                  Reset
                </button>
              )}
            </form>

            {uploadProgress > 0 && (
              <div className="mt-4 text-gray-700 dark:text-gray-300">
                Upload Progress: {uploadProgress}%
              </div>
            )}

            {status && (
              <p className="mt-2 text-gray-700 dark:text-gray-300">{status}</p>
            )}

            {videoUrl && (
              <div className="mt-6">
                <video
                  ref={videoRef}
                  controls
                  className="w-full rounded-lg"
                  preload="metadata"
                />
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
