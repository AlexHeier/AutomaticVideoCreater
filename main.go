package main

import (
	"fmt"
	"log"
	"os"
	editVideo "videoCreater/editVideo"
	getVideo "videoCreater/getVideo"
	global "videoCreater/global"
	quote "videoCreater/quote"
	upload "videoCreater/upload"
	voice "videoCreater/voice"

	"github.com/joho/godotenv"
)

func init() {
	initConfig()
}

func main() {
	CreateQuoteVideo()
}

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
}

// Initialize environment and verify configuration
func initConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}

	// Check if the API key is set
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		log.Fatalf("GOOGLE_API_KEY environment variable is not set")
	}

	// Ensure necessary directories exist
	ensureDirExists("raw-videos")
	ensureDirExists("edited-videos")
	ensureDirExists("text-to-speeched")
}

// Ensure directory exists
func ensureDirExists(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.Mkdir(dir, 0755); err != nil {
			log.Fatalf("Failed to create directory %s: %v", dir, err)
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
