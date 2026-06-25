package transcoder

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"lms-video-worker/internal/azureblob"
)

// ProcessAdaptiveStreaming takes a raw video URL, runs ffmpeg to generate HLS streams, and uploads them to Azure Blob Storage
func ProcessAdaptiveStreaming(tenantID, videoID uint, rawUrl string) (string, error) {
	log.Printf("Starting Adaptive Streaming Transcode for Tenant %d, Video %d", tenantID, videoID)
	
	// Create temporary working directory for this job
	workDir := fmt.Sprintf("/tmp/hls/%d/%d", tenantID, videoID)
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create workdir: %w", err)
	}
	defer os.RemoveAll(workDir) // Cleanup after upload

	// We'll generate 3 qualities: 1080p, 720p, 480p
	// In a real production scenario, you would check the original resolution first
	
	// 1. Download raw video or stream it directly into ffmpeg. 
	// NOTE: If rawUrl is an Azure SAS URL, ffmpeg can read it directly.

	// 2. FFMPEG Command for HLS
	// This splits the video into 10-second .ts chunks and generates the playlist
	cmd := exec.Command("ffmpeg",
		"-hide_banner", "-y",
		"-i", rawUrl,
		
		// 1080p
		"-map", "0:v:0", "-map", "0:a:0",
		"-s:v:0", "1920x1080", "-c:v:0", "libx264", "-b:v:0", "5000k",
		
		// 720p
		"-map", "0:v:0", "-map", "0:a:0",
		"-s:v:1", "1280x720", "-c:v:1", "libx264", "-b:v:1", "2800k",
		
		// 480p
		"-map", "0:v:0", "-map", "0:a:0",
		"-s:v:2", "854x480", "-c:v:2", "libx264", "-b:v:2", "1400k",
		
		// Audio config
		"-c:a", "aac", "-b:a", "128k",

		// HLS config
		"-f", "hls",
		"-hls_time", "10",
		"-hls_playlist_type", "vod",
		"-master_pl_name", "master.m3u8",
		"-var_stream_map", "v:0,a:0 v:1,a:1 v:2,a:2",
		filepath.Join(workDir, "stream_%v.m3u8"),
	)
	_ = cmd

	// 3. Run FFMPEG
	log.Printf("Running FFMPEG: %s", cmd.String())
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("ffmpeg failed: %w", err)
	}

	// 4. Upload all generated .ts and .m3u8 files to Azure Blob Storage
	log.Printf("Generated HLS files in %s. Uploading to Azure Blob Container...", workDir)

	files, err := os.ReadDir(workDir)
	if err != nil {
		return "", fmt.Errorf("failed to read output directory: %w", err)
	}

	var masterUrl string
	for _, f := range files {
		localPath := filepath.Join(workDir, f.Name())
		blobName := fmt.Sprintf("%d/videos/%d/hls/%s", tenantID, videoID, f.Name())
		
		url, err := azureblob.UploadFile(localPath, blobName)
		if err != nil {
			log.Printf("Failed to upload %s: %v", f.Name(), err)
			return "", err
		}
		log.Printf("Uploaded: %s", url)

		if f.Name() == "master.m3u8" {
			masterUrl = url
		}
	}

	// 5. Emit Transcoded Event to Kafka - MOVED TO CONSUMER
	log.Println("Transcoding and upload complete!")
	return masterUrl, nil
}
