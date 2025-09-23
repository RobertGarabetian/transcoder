"use client"; // only if using App Router (app/page.tsx)

import { useState } from "react";

export default function Home() {
  const [file, setFile] = useState<File | null>(null);
  const [status, setStatus] = useState("");

  const handleUpload = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!file) {
      setStatus("Please choose a file first.");
      return;
    }

    const formData = new FormData();
    formData.append("video", file);

    try {
      const res = await fetch("http://localhost:8080/upload", {
        method: "POST",
        body: formData,
      });

      if (res.ok) {
        const text = await res.text();
        setStatus("✅ " + text);
      } else {
        setStatus("❌ Upload failed: " + res.statusText);
      }
    } catch (err) {
      setStatus("❌ Error: " + (err as Error).message);
    }
  };

  return (
    <main style={{ padding: "2rem", fontFamily: "sans-serif" }}>
      <h1>Upload a Video</h1>
      <form onSubmit={handleUpload}>
        <input
          type="file"
          accept="video/*"
          onChange={(e) => setFile(e.target.files?.[0] || null)}
        />
        <button type="submit" style={{ marginLeft: "1rem" }}>
          Upload
        </button>
      </form>
      {status && <p>{status}</p>}
    </main>
  );
}
