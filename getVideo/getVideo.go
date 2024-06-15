package getVideo

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// PexelsVideoResponse represents the response structure from Pexels API
type PexelsVideoResponse struct {
	Videos []struct {
		Id         int `json:"id"`
		VideoFiles []struct {
			Link   string `json:"link"`
			Width  int    `json:"width"`
			Height int    `json:"height"`
		} `json:"video_files"`
	} `json:"videos"`
}

// FetchAndStoreVideo fetches a video from Pexels based on the theme and stores it in the raw-videos folder
func FetchAndStoreVideo(thema string) (string, error) {
	apiKey := os.Getenv("PEXELS_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("PEXELS_API_KEY environment variable is not set")
	}

	maxRetries := 5
	for attempt := 1; attempt <= maxRetries; attempt++ {
		// Request more videos by increasing per_page parameter
		url := fmt.Sprintf("https://api.pexels.com/videos/search?query=%s&per_page=10", thema)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return "", fmt.Errorf("failed to create request: %v", err)
		}
		req.Header.Set("Authorization", apiKey)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return "", fmt.Errorf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read response body: %v", err)
		}

		var videoResp PexelsVideoResponse
		err = json.Unmarshal(body, &videoResp)
		if err != nil {
			return "", fmt.Errorf("failed to parse response body: %v", err)
		}

		if len(videoResp.Videos) == 0 {
			log.Printf("No videos found for theme: %s. Retrying...", thema)
			time.Sleep(3 * time.Second)
			continue
		}

		// Select a random video from the results
		rand.Seed(time.Now().UnixNano())
		randomIndex := rand.Intn(len(videoResp.Videos))
		selectedVideo := videoResp.Videos[randomIndex]

		var videoURL string
		for _, file := range selectedVideo.VideoFiles {
			if file.Width == 1920 && file.Height == 1080 {
				videoURL = file.Link
				break
			}
		}

		if videoURL == "" {
			log.Printf("No video found with resolution 1920x1080 for theme: %s. Retrying...", thema)
			time.Sleep(3 * time.Second)
			continue
		}

		// Ensure the directory exists
		dir := "raw-videos"
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err := os.Mkdir(dir, 0755); err != nil {
				return "", fmt.Errorf("failed to create directory: %v", err)
			}
		}

		videoPath := filepath.Join(dir, fmt.Sprintf("%d.mp4", selectedVideo.Id))

		// Download the video
		err = downloadFile(videoPath, videoURL)
		if err != nil {
			return "", fmt.Errorf("failed to download video: %v", err)
		}

		log.Printf("Video downloaded to: %s\n", videoPath)
		return videoPath, nil
	}

	return "", fmt.Errorf("failed to find a suitable video after %d attempts", maxRetries)
}

// downloadFile downloads a file from the given URL and saves it to the specified path
func downloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to copy content: %v", err)
	}

	return nil
}
