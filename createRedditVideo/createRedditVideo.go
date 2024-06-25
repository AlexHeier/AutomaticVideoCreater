package createredditvideo

import (
	"fmt"
	"log"
	"os"
	"time"
	"videoCreater/createRedditVideo/reddit"
	"videoCreater/editVideo"
	"videoCreater/getVideo"
	"videoCreater/global"
	upload "videoCreater/upload"
	"videoCreater/voice"
)

func CreateRedditVideo(subreddit string) {

	// Create the video
	outputVideoPath, err := createVideo(subreddit)
	if err != nil {
		log.Fatalf("Failed to create video: %v", err)
	}
	if global.DeleteTikTokVideoAfterPost {
		defer removeFiles(outputVideoPath)
	}

	if global.PostTikTokVideo {
		// Ensure outputVideoPath is accessible and has contents
		for i := 0; i < len(outputVideoPath); i++ {
			currentDate := time.Now().Format("02.01.2006") // Correct date format
			title := ""
			if len(outputVideoPath) > 1 {
				title = fmt.Sprintf("Reddit: %s part %d of %d - %s", subreddit, i+1, len(outputVideoPath), currentDate)
			} else {
				title = fmt.Sprintf("Reddit: %s - %s", subreddit, currentDate)
			}

			description := fmt.Sprintf("%v? Leave a comment about what you think!", subreddit)

			err := upload.UploadVideoTikTok(outputVideoPath[i], title, description)
			if err != nil {
				log.Printf("Failed to upload video: %v", err)
			}
		}
	}
}

// Create the video
func createVideo(subreddit string) ([]string, error) {

	// Fetch quote
	post, err := reddit.GetRedditPost(subreddit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch reddit post: %v", err)
	}

	text := fmt.Sprintf("%v %v", post.Title, post.Content)

	// Convert text to speech
	pathToVoice, wordTimings, err := voice.ConvertTextToSpeech(text)
	if err != nil {
		return nil, fmt.Errorf("failed to convert text to speech: %v", err)
	}
	defer removeFiles(pathToVoice)

	// Fetch video
	pathToVideo, err := getVideo.FetchAndDownloadYoutubeVideo("subway surfers gameplay no copyright", (3*len(pathToVoice))+1, (10*len(pathToVoice))+1)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch video: %v", err)
	}
	defer os.Remove(pathToVideo)

	outputVideoPath, err := editVideo.EditVideoTikTok(pathToVideo, pathToVoice, wordTimings, post.Title)
	if err != nil {
		return nil, fmt.Errorf("failed to edit video: %v", err)
	}

	return outputVideoPath, nil
}

func removeFiles(paths []string) {
	for _, path := range paths {
		err := os.Remove(path)
		if err != nil {
			fmt.Printf("Failed to remove file %s: %v\n", path, err)
		}
	}
}
