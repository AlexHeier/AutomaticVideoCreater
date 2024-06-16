package createquotevideo

import (
	"fmt"
	"log"
	"os"
	"videoCreater/createQuoteVideo/editVideo"
	"videoCreater/createQuoteVideo/quote"
	"videoCreater/getVideo"
	"videoCreater/global"
	"videoCreater/upload"
	"videoCreater/voice"
)

func CreateQuoteVideo() {
	thema := global.Random(global.Themas)
	// Create the video
	outputVideoPath, err := createVideo(thema)
	if err != nil {
		log.Fatalf("Failed to create video: %v", err)
	}

	if global.PostVideo {
		err = upload.UploadVideo(outputVideoPath, thema)
		if err != nil {
			log.Fatalf("Failed to upload video: %v", err)
		}
	}
	if global.DeleteVideoAfterPost {
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
	pathToVoice, err := voice.ConvertTextToSpeech(content)
	if err != nil {
		return "", fmt.Errorf("failed to convert text to speech: %v", err)
	}
	defer os.Remove(pathToVoice)

	// Fetch video
	pathToVideo, err := getVideo.FetchAndStoreVideo(thema)
	if err != nil {
		return "", fmt.Errorf("failed to fetch video: %v", err)
	}
	defer os.Remove(pathToVideo)

	// Edit video
	outputVideoPath, err := editVideo.EditVideo(pathToVideo, pathToVoice, thema, content, author)
	if err != nil {
		return "", fmt.Errorf("failed to edit video: %v", err)
	}

	return outputVideoPath, nil
}
