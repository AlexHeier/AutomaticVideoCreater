package main

import (
	"log"
	"os"

	createQuoteVideo "videoCreater/createQuoteVideo"

	"github.com/joho/godotenv"
)

func init() {
	initConfig()
}

func main() {
	createQuoteVideo.CreateQuoteVideo()
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
