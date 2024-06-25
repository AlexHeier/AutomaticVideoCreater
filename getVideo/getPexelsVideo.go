package getVideo

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
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

// FetchAndStoreVideos fetches a specified number of videos from Pexels based on the theme and stores them in the raw-videos folder
func FetchAndStoreVideosPexels(theme string, amount int) ([]string, error) {
	apiKey := os.Getenv("PEXELS_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("PEXELS_API_KEY environment variable is not set")
	}

	var videoPaths []string
	seenIDs := make(map[int]bool) // To avoid downloading duplicates

	for len(videoPaths) < amount {
		url := fmt.Sprintf("https://api.pexels.com/videos/search?query=%s&per_page=%d", theme, amount*2) // Request more to account for filtering
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %v", err)
		}
		req.Header.Set("Authorization", apiKey)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %v", err)
		}

		var videoResp PexelsVideoResponse
		err = json.Unmarshal(body, &videoResp)
		if err != nil {
			return nil, fmt.Errorf("failed to parse response body: %v", err)
		}

		if len(videoResp.Videos) == 0 {
			log.Printf("No more videos found for theme: %s.", theme)
			break
		}

		for _, video := range videoResp.Videos {
			if seenIDs[video.Id] {
				continue // Skip if already processed this video
			}
			seenIDs[video.Id] = true

			for _, file := range video.VideoFiles {
				if file.Width == 1920 && file.Height == 1080 {
					videoPath, err := downloadAndSaveVideo(file.Link, video.Id)
					if err != nil {
						log.Printf("Failed to download video: %v", err)
						continue
					}
					videoPaths = append(videoPaths, videoPath)
					if len(videoPaths) >= amount {
						break // Stop if we have enough videos
					}
				}
			}
		}
	}

	if len(videoPaths) < amount {
		return videoPaths, fmt.Errorf("only found %d videos out of requested %d", len(videoPaths), amount)
	}
	return videoPaths, nil
}

func downloadAndSaveVideo(url string, id int) (string, error) {
	dir := "raw-videos"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.Mkdir(dir, 0755); err != nil {
			return "", fmt.Errorf("failed to create directory: %v", err)
		}
	}

	videoPath := filepath.Join(dir, fmt.Sprintf("%d.mp4", id))
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	out, err := os.Create(videoPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to copy content: %v", err)
	}

	return videoPath, nil
}
