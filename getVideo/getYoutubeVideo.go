package getVideo

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kkdai/youtube/v2"
)

// YouTubeSearchResponse represents the response structure from YouTube API
type YouTubeSearchResponse struct {
	Items []struct {
		ID struct {
			VideoID string `json:"videoId"`
		} `json:"id"`
	} `json:"items"`
}

// YouTubeVideoDetails represents the details of a YouTube video
type YouTubeVideoDetails struct {
	Items []struct {
		ContentDetails struct {
			Duration string `json:"duration"`
		} `json:"contentDetails"`
	} `json:"items"`
}

// FetchYoutubeVideos searches for videogame gameplay videos on YouTube and returns their video URLs
func FetchYoutubeVideos(query string, maxResults int) ([]string, error) {
	apiKey := os.Getenv("YOUTUBE_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("YOUTUBE_API_KEY environment variable is not set")
	}

	encodedQuery := url.QueryEscape(query)
	url := fmt.Sprintf("https://www.googleapis.com/youtube/v3/search?part=id&type=video&q=%s&maxResults=%d&key=%s", encodedQuery, maxResults, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error fetching videos: status code %d, body: %s", resp.StatusCode, string(body))
	}

	var searchResponse YouTubeSearchResponse
	if err := json.Unmarshal(body, &searchResponse); err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %v, body: %s", err, string(body))
	}

	var videoURLs []string
	for _, item := range searchResponse.Items {
		videoID := item.ID.VideoID
		videoURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)
		videoURLs = append(videoURLs, videoURL)
	}

	return videoURLs, nil
}

// getVideoDuration fetches the duration of a video by its ID
func getVideoDuration(videoID string) (time.Duration, error) {
	apiKey := os.Getenv("YOUTUBE_API_KEY")
	if apiKey == "" {
		return 0, fmt.Errorf("YOUTUBE_API_KEY environment variable is not set")
	}

	url := fmt.Sprintf("https://www.googleapis.com/youtube/v3/videos?part=contentDetails&id=%s&key=%s", videoID, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("error fetching video details: status code %d, body: %s", resp.StatusCode, string(body))
	}

	var details YouTubeVideoDetails
	if err := json.Unmarshal(body, &details); err != nil {
		return 0, fmt.Errorf("error unmarshalling response: %v, body: %s", err, string(body))
	}

	if len(details.Items) == 0 {
		return 0, fmt.Errorf("no video details found for video ID %s", videoID)
	}

	durationStr := details.Items[0].ContentDetails.Duration
	duration, err := parseISO8601Duration(durationStr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse video duration: %v", err)
	}

	return duration, nil
}

// parseISO8601Duration parses an ISO 8601 duration string and returns the duration in seconds
func parseISO8601Duration(isoDuration string) (time.Duration, error) {
	isoDuration = strings.ToLower(isoDuration)
	isoDuration = strings.Replace(isoDuration, "pt", "", 1)
	isoDuration = strings.Replace(isoDuration, "h", "h", 1)
	isoDuration = strings.Replace(isoDuration, "m", "m", 1)
	isoDuration = strings.Replace(isoDuration, "s", "s", 1)
	return time.ParseDuration(isoDuration)
}

// DownloadVideo downloads a video from YouTube and saves it locally
func DownloadVideo(videoURL string) (string, error) {
	client := youtube.Client{}

	video, err := client.GetVideo(videoURL)
	if err != nil {
		return "", fmt.Errorf("failed to get video info: %v", err)
	}

	saveDir := "raw-videos"
	if _, err := os.Stat(saveDir); os.IsNotExist(err) {
		if err := os.Mkdir(saveDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create directory: %v", err)
		}
	}

	filePath := filepath.Join(saveDir, fmt.Sprintf("%s.mp4", video.ID))
	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	stream, _, err := client.GetStream(video, &video.Formats[0])
	if err != nil {
		return "", fmt.Errorf("failed to get video stream: %v", err)
	}

	_, err = io.Copy(file, stream)
	if err != nil {
		return "", fmt.Errorf("failed to save video: %v", err)
	}

	return filePath, nil
}

// FetchAndDownloadYoutubeVideo fetches and downloads a single gameplay video from YouTube that fits the duration range
func FetchAndDownloadYoutubeVideo(query string, minDuration, maxDuration int) (string, error) {
	const maxRetries = 5
	var lastError error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		videoURLs, err := FetchYoutubeVideos(query, 50) // Fetch more videos to ensure we get enough within the duration range
		if err != nil {
			lastError = err
			log.Printf("Attempt %d: Failed to fetch YouTube videos: %v", attempt, err)
			time.Sleep(time.Second * time.Duration(attempt)) // Exponential backoff
			continue
		}

		var suitableVideos []string
		for _, videoURL := range videoURLs {
			videoID := strings.TrimPrefix(videoURL, "https://www.youtube.com/watch?v=")
			duration, err := getVideoDuration(videoID)
			if err != nil {
				log.Printf("Attempt %d: Failed to get duration for video %s: %v", attempt, videoURL, err)
				continue
			}

			if duration >= time.Duration(minDuration)*time.Minute && duration <= time.Duration(maxDuration)*time.Minute {
				suitableVideos = append(suitableVideos, videoURL)
			}
		}

		if len(suitableVideos) == 0 {
			lastError = fmt.Errorf("no videos found within the specified duration range on attempt %d", attempt)
			log.Printf("Attempt %d: No suitable videos found", attempt)
			time.Sleep(time.Second * time.Duration(attempt)) // Exponential backoff
			continue
		}

		// Select a random video from the suitable videos
		rand.Seed(time.Now().UnixNano())
		selectedVideoURL := suitableVideos[rand.Intn(len(suitableVideos))]

		videoPath, err := DownloadVideo(selectedVideoURL)
		if err != nil {
			lastError = err
			log.Printf("Attempt %d: Failed to download video %s: %v", attempt, selectedVideoURL, err)
			time.Sleep(time.Second * time.Duration(attempt)) // Exponential backoff
			continue
		}

		return videoPath, nil
	}

	return "", lastError
}
