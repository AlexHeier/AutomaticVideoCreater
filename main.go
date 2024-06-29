package main

import (
	"log"
	"os"

	createQuoteVideo "videoCreater/createQuoteVideo"
	createRedditVideo "videoCreater/createRedditVideo"

	"github.com/joho/godotenv"
)

func init() {
	initConfig()
}

func main() {
	// Checks if the minimum number of arguments is not met
	if len(os.Args) < 2 {
		log.Println("Usage: videocreater <videotype> [extra info]")
		os.Exit(1)
	}
	videoType := os.Args[1]

	switch videoType {
	case "quote":
		createQuoteVideo.CreateQuoteVideo()

	case "reddit":
		// Ensure that subreddit argument is also provided
		if len(os.Args) < 3 {
			log.Println("videotype 'reddit' requires the subreddit name as well")
			os.Exit(1)
		}
		createRedditVideo.CreateRedditVideo(os.Args[2])

	default:
		log.Println("Unknown videotype. Use 'quote' or 'reddit'.")
		os.Exit(1) // Exit after logging the unknown type error
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
