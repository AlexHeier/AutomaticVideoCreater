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

	speechpb "google.golang.org/genproto/googleapis/cloud/speech/v1"
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
	var wordTimings []*speechpb.WordInfo
	// Fetch quote
	content, author, err := quote.FetchQuote(thema)
	if err != nil {
		return "", fmt.Errorf("failed to fetch quote: %v", err)
	}

	// Convert text to speech
	pathToVoice, wordTimings, err := voice.ConvertTextToSpeech(content)
	if err != nil {
		return "", fmt.Errorf("failed to convert text to speech: %v", err)
	}

	// Fetch video
	pathToVideo, err := getVideo.FetchAndStoreVideo(thema)
	if err != nil {
		return "", fmt.Errorf("failed to fetch video: %v", err)
	}

	// Edit video
	outputVideoPath, err := editVideo.EditVideo(pathToVideo, pathToVoice, wordTimings, thema, content, author)
	if err != nil {
		return "", fmt.Errorf("failed to edit video: %v", err)
	}

	if global.DeleteVideoParts {
		fmt.Println("Entering delete video parts section")
		err = os.Remove(pathToVoice)
		if err != nil {
			fmt.Printf("failed to delete voice file: %v\n", err)
		}

		err = os.Remove(pathToVideo)
		if err != nil {
			fmt.Printf("failed to delete video file: %v\n", err)
		}

		err = os.Remove("text-to-speeched/converted_audio.wav")
		if err != nil {
			fmt.Printf("failed to delete converted audio file: %v\n", err)
		}
	}

	return outputVideoPath, nil
}
