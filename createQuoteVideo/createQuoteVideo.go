package createquotevideo

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
	"videoCreater/createQuoteVideo/quote"
	"videoCreater/editVideo"
	"videoCreater/getVideo"
	"videoCreater/global"
	"videoCreater/upload"
	"videoCreater/voice"
)

func CreateQuoteVideo() {
	thema := random(global.Themas)
	// Create the video
	outputVideoPath, err := createVideo(thema)
	if err != nil {
		log.Fatalf("Failed to create video: %v", err)
	}

	title := fmt.Sprintf("A Quote of %s", strings.Title(thema))
	description := fmt.Sprintf("A beautiful quote about %s. Leave a Like and Subscribe for more beautiful quotes ", strings.Title(thema))
	chategoryID := "22"

	if global.PostYoutubeVideo {
		err = upload.UploadVideoYoutube(outputVideoPath, description, title, chategoryID, global.Themas)
		if err != nil {
			log.Fatalf("Failed to upload video: %v", err)
		}
	}
	if global.DeleteYoutubeVideoAfterPost {
		err = os.Remove(outputVideoPath)
		if err != nil {
			log.Printf("Error deleting video file: %v", err)
		}
	}
}

// Create the video
func createVideo(thema string) (string, error) {
	// Fetch quote
	content, author, err := quote.FetchQuote(thema)
	if err != nil {
		return "", fmt.Errorf("failed to fetch quote: %v", err)
	}

	// Convert text to speech
	pathToVoices, wordTimings, err := voice.ConvertTextToSpeech(content)
	if err != nil {
		return "", fmt.Errorf("failed to convert text to speech: %v", err)
	}
	defer removeFiles(pathToVoices)

	// Fetch video
	pathToVideos, err := getVideo.FetchAndStoreVideosPexels(thema, len(pathToVoices))
	if err != nil {
		return "", fmt.Errorf("failed to fetch video: %v", err)
	}
	defer removeFiles(pathToVideos)

	// Edit video
	var title string = fmt.Sprintf("A Quote of %s", strings.Title(thema))

	outputVideoPath, err := editVideo.EditVideoYoutube(pathToVideos[0], pathToVoices[0], wordTimings[0], title, author)
	if err != nil {
		return "", fmt.Errorf("failed to edit video: %v", err)
	}

	return outputVideoPath, nil
}

func random(array []string) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return array[r.Intn(len(array))]
}

func removeFiles(paths []string) {
	for _, path := range paths {
		err := os.Remove(path)
		if err != nil {
			fmt.Printf("Failed to remove file %s: %v\n", path, err)
		}
	}
}
