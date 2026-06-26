package transcoder
import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"path/filepath"
	"lms-video-worker/internal/azureblob"
)

func ProcessAdaptiveStreaming(tenantID, courseID, videoID uint, rawUrl, slug, phone string) (string, error) {
	log.Printf("Starting Adaptive Streaming Transcode for Tenant %d, Course %d, Video %d", tenantID, courseID, videoID)
	workDir := fmt.Sprintf("/tmp/hls/%d/%d", tenantID, videoID)
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create workdir: %w", err)
	}
	defer os.RemoveAll(workDir)
	ffmpegInput := rawUrl
	if strings.HasPrefix(rawUrl, "local://") {
		ffmpegInput = strings.TrimPrefix(rawUrl, "local://")
	} else {
		ffmpegInput = strings.ReplaceAll(ffmpegInput, " ", "%20")
	}

	args := []string{
		"-hide_banner", "-nostdin", "-y",
		"-i", ffmpegInput,
		"-preset", "veryfast",
		"-g", "48", "-keyint_min", "48", "-sc_threshold", "0",
		"-map", "0:v:0", "-map", "0:a:0",
		"-s:v:0", "1920x1080", "-c:v:0", "libx264", "-b:v:0", "5000k",
		"-map", "0:v:0", "-map", "0:a:0",
		"-s:v:1", "1280x720", "-c:v:1", "libx264", "-b:v:1", "2800k",
		"-map", "0:v:0", "-map", "0:a:0",
		"-s:v:2", "854x480", "-c:v:2", "libx264", "-b:v:2", "1400k",
		"-c:a", "aac", "-b:a", "128k",
		"-f", "hls",
		"-hls_time", "4",
		"-hls_playlist_type", "vod",
		"-master_pl_name", "master.m3u8",
		"-var_stream_map", "v:0,a:0 v:1,a:1 v:2,a:2",
		filepath.Join(workDir, "stream_%v.m3u8"),
	}

	cmd := exec.Command("ffmpeg", args...)
	cmd.Dir = workDir
	log.Printf("Running FFMPEG with audio: %s", cmd.String())
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("FFMPEG with audio failed. Output: %s", string(out))

		log.Printf("Retrying FFMPEG without audio streams...")
		argsNoAudio := []string{
			"-hide_banner", "-nostdin", "-y",
			"-i", ffmpegInput,
			"-preset", "veryfast",
			"-g", "48", "-keyint_min", "48", "-sc_threshold", "0",
			"-map", "0:v:0",
			"-s:v:0", "1920x1080", "-c:v:0", "libx264", "-b:v:0", "5000k",
			"-map", "0:v:0",
			"-s:v:1", "1280x720", "-c:v:1", "libx264", "-b:v:1", "2800k",
			"-map", "0:v:0",
			"-s:v:2", "854x480", "-c:v:2", "libx264", "-b:v:2", "1400k",
			"-f", "hls",
			"-hls_time", "4",
			"-hls_playlist_type", "vod",
			"-master_pl_name", "master.m3u8",
			"-var_stream_map", "v:0 v:1 v:2",
			filepath.Join(workDir, "stream_%v.m3u8"),
		}

		cmdNoAudio := exec.Command("ffmpeg", argsNoAudio...)
		cmdNoAudio.Dir = workDir
		outNoAudio, errNoAudio := cmdNoAudio.CombinedOutput()
		if errNoAudio != nil {
			return "", fmt.Errorf("ffmpeg fallback failed: %w, output: %s", errNoAudio, string(outNoAudio))
		}
	}

	log.Printf("Generated HLS files in %s. Uploading to Azure Blob Container...", workDir)
	files, err := os.ReadDir(workDir)
	if err != nil {
		return "", fmt.Errorf("failed to read output directory: %w", err)
	}
	var masterUrl string
	for _, f := range files {
		localPath := filepath.Join(workDir, f.Name())
		blobName := fmt.Sprintf("%s_%s/video/%d/%d/hls/%s", slug, phone, courseID, videoID, f.Name())
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
	log.Println("Transcoding and upload complete!")
	return masterUrl, nil
}
