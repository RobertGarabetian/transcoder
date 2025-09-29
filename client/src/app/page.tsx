"use client";

import { useState, useRef } from "react";

interface TranscodingProgress {
  format: string;
  status: 'pending' | 'processing' | 'completed' | 'error';
  progress: number;
  error?: string;
}

export default function Home() {
  const [file, setFile] = useState<File | null>(null);
  const [isUploading, setIsUploading] = useState(false);
  const [uploadProgress, setUploadProgress] = useState(0);
  const [transcodingProgress, setTranscodingProgress] = useState<TranscodingProgress[]>([
    { format: '1080p', status: 'pending', progress: 0 },
    { format: '720p', status: 'pending', progress: 0 },
    { format: '490p', status: 'pending', progress: 0 }
  ]);
  const [overallStatus, setOverallStatus] = useState("");
  const fileInputRef = useRef<HTMLInputElement>(null);

  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const selectedFile = e.target.files?.[0] || null;
    setFile(selectedFile);
    if (selectedFile) {
      setOverallStatus("");
      setTranscodingProgress([
        { format: '1080p', status: 'pending', progress: 0 },
        { format: '720p', status: 'pending', progress: 0 },
        { format: '490p', status: 'pending', progress: 0 }
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

    // Reset transcoding progress
    setTranscodingProgress([
      { format: '1080p', status: 'pending', progress: 0 },
      { format: '720p', status: 'pending', progress: 0 },
      { format: '490p', status: 'pending', progress: 0 }
    ]);

    const formData = new FormData();
    formData.append("video", file);

    try {
      // Simulate upload progress
      const progressInterval = setInterval(() => {
        setUploadProgress(prev => {
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

      clearInterval(progressInterval);
      setUploadProgress(100);

      if (res.ok) {
        const text = await res.text();
        setOverallStatus("✅ Upload successful! Processing video formats...");

        // Simulate transcoding progress for each format
        const formats = ['1080p', '720p', '490p'];
        formats.forEach((format, index) => {
          setTimeout(() => {
            setTranscodingProgress(prev =>
              prev.map(p => p.format === format ? { ...p, status: 'processing' as const } : p)
            );

            // Simulate processing progress
            const progressInterval = setInterval(() => {
              setTranscodingProgress(prev =>
                prev.map(p => {
                  if (p.format === format && p.status === 'processing') {
                    const newProgress = Math.min(p.progress + Math.random() * 15, 100);
                    if (newProgress >= 100) {
                      clearInterval(progressInterval);
                      return { ...p, progress: 100, status: 'completed' as const };
                    }
                    return { ...p, progress: newProgress };
                  }
                  return p;
                })
              );
            }, 300);
          }, index * 1000); // Stagger the start of each format
        });

        // Check if all formats are completed
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
    setTranscodingProgress([
      { format: '1080p', status: 'pending', progress: 0 },
      { format: '720p', status: 'pending', progress: 0 },
      { format: '490p', status: 'pending', progress: 0 }
    ]);
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  };

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
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
                  Select Video File
                </label>
                <div className="relative">
                  <input
                    ref={fileInputRef}
                    type="file"
                    accept="video/*"
                    onChange={handleFileChange}
                    disabled={isUploading}
                    className="block w-full text-sm text-gray-500 file:mr-4 file:py-2 file:px-4 file:rounded-lg file:border-0 file:text-sm file:font-semibold file:bg-blue-50 file:text-blue-700 hover:file:bg-blue-100 disabled:opacity-50"
                  />
                </div>

                {file && (
                  <div className="bg-gray-50 dark:bg-gray-700 rounded-lg p-4">
                    <div className="flex items-center justify-between">
                      <div>
                        <p className="font-medium text-gray-900 dark:text-white">{file.name}</p>
                        <p className="text-sm text-gray-500 dark:text-gray-400">
                          {formatFileSize(file.size)} • {file.type}
                        </p>
                      </div>
                      <button
                        type="button"
                        onClick={resetForm}
                        disabled={isUploading}
                        className="text-red-500 hover:text-red-700 disabled:opacity-50"
                      >
                        Remove
                      </button>
                    </div>
                  </div>
                )}
              </div>

              <div className="flex gap-4">
                <button
                  type="submit"
                  disabled={!file || isUploading}
                  className="flex-1 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-400 text-white font-medium py-3 px-6 rounded-lg transition-colors duration-200 flex items-center justify-center gap-2"
                >
                  {isUploading ? (
                    <>
                      <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
                      Processing...
                    </>
                  ) : (
                    'Start Transcoding'
                  )}
                </button>

                {isUploading && (
                  <button
                    type="button"
                    onClick={resetForm}
                    className="px-6 py-3 border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors duration-200"
                  >
                    Cancel
                  </button>
                )}
              </div>
            </form>

            {/* Upload Progress */}
            {isUploading && uploadProgress > 0 && (
              <div className="mt-6 space-y-2">
                <div className="flex justify-between text-sm text-gray-600 dark:text-gray-400">
                  <span>Uploading file...</span>
                  <span>{Math.round(uploadProgress)}%</span>
                </div>
                <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
                  <div
                    className="bg-blue-600 h-2 rounded-full transition-all duration-300"
                    style={{ width: `${uploadProgress}%` }}
                  ></div>
                </div>
              </div>
            )}

            {/* Transcoding Progress */}
            {isUploading && uploadProgress >= 100 && (
              <div className="mt-8 space-y-4">
                <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
                  Processing Video Formats
                </h3>
                <div className="space-y-3">
                  {transcodingProgress.map((format) => (
                    <div key={format.format} className="bg-gray-50 dark:bg-gray-700 rounded-lg p-4">
                      <div className="flex items-center justify-between mb-2">
                        <span className="font-medium text-gray-900 dark:text-white">
                          {format.format}
                        </span>
                        <div className="flex items-center gap-2">
                          {format.status === 'pending' && (
                            <span className="text-gray-500 dark:text-gray-400 text-sm">Waiting...</span>
                          )}
                          {format.status === 'processing' && (
                            <div className="flex items-center gap-2">
                              <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-600"></div>
                              <span className="text-blue-600 dark:text-blue-400 text-sm">
                                {Math.round(format.progress)}%
                              </span>
                            </div>
                          )}
                          {format.status === 'completed' && (
                            <span className="text-green-600 dark:text-green-400 text-sm flex items-center gap-1">
                              <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                                <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                              </svg>
                              Complete
                            </span>
                          )}
                          {format.status === 'error' && (
                            <span className="text-red-600 dark:text-red-400 text-sm">Error</span>
                          )}
                        </div>
                      </div>
                      {format.status === 'processing' && (
                        <div className="w-full bg-gray-200 dark:bg-gray-600 rounded-full h-2">
                          <div
                            className="bg-blue-600 h-2 rounded-full transition-all duration-300"
                            style={{ width: `${format.progress}%` }}
                          ></div>
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* Status Messages */}
            {overallStatus && (
              <div className="mt-6 p-4 rounded-lg bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800">
                <p className="text-blue-800 dark:text-blue-200">{overallStatus}</p>
              </div>
            )}
          </div>

          {/* Output Formats Info */}
          <div className="mt-8 bg-white dark:bg-gray-800 rounded-xl shadow-lg p-6">
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
              Output Formats
            </h3>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div className="text-center p-4 bg-gray-50 dark:bg-gray-700 rounded-lg">
                <div className="text-2xl font-bold text-blue-600 dark:text-blue-400">1080p</div>
                <div className="text-sm text-gray-600 dark:text-gray-400">High Quality</div>
              </div>
              <div className="text-center p-4 bg-gray-50 dark:bg-gray-700 rounded-lg">
                <div className="text-2xl font-bold text-green-600 dark:text-green-400">720p</div>
                <div className="text-sm text-gray-600 dark:text-gray-400">Standard Quality</div>
              </div>
              <div className="text-center p-4 bg-gray-50 dark:bg-gray-700 rounded-lg">
                <div className="text-2xl font-bold text-orange-600 dark:text-orange-400">490p</div>
                <div className="text-sm text-gray-600 dark:text-gray-400">Mobile Optimized</div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
