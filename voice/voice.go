package voice

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"videoCreater/global"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	texttospeechpb "cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"google.golang.org/api/option"
)

// findNextAvailableFilename finds the lowest available number for the filename in the specified directory.
func findNextAvailableFilename(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}

	occupied := make(map[int]bool)
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, "voice") && strings.HasSuffix(name, ".mp3") {
			numStr := strings.TrimSuffix(strings.TrimPrefix(name, "voice"), ".mp3")
			num, err := strconv.Atoi(numStr)
			if err == nil {
				occupied[num] = true
			}
		}
	}

	// Find the lowest available number
	i := 1
	for ; occupied[i]; i++ {
	}
	return filepath.Join(dir, "voice"+strconv.Itoa(i)+".mp3"), nil
}

// ConvertTextToSpeech takes a string and converts the text to speech, saving the output as an MP3 file
// in the "text-to-speeched" directory and returning the path to the created file.
func ConvertTextToSpeech(text string) (string, error) {
	ctx := context.Background()

	// Log the GOOGLE_API_KEY environment variable
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("GOOGLE_API_KEY environment variable is not set")
	}
	log.Printf("Using API key: %s", apiKey)

	// Initialize the Google Cloud Text-to-Speech client with the API key
	client, err := texttospeech.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return "", fmt.Errorf("failed to create client: %v", err)
	}
	defer client.Close()

	// Ensure the directory exists
	dir := "text-to-speeched"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.Mkdir(dir, 0755); err != nil {
			return "", fmt.Errorf("failed to create directory: %v", err)
		}
	}

	fullPath, err := findNextAvailableFilename(dir)
	if err != nil {
		return "", fmt.Errorf("failed to find next available filename: %v", err)
	}

	log.Println("Setting up the SynthesizeSpeech request...")
	req := &texttospeechpb.SynthesizeSpeechRequest{
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{Text: text},
		},
		Voice: &texttospeechpb.VoiceSelectionParams{
			LanguageCode: "en-US",
			SsmlGender:   texttospeechpb.SsmlVoiceGender_NEUTRAL,
		},
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding: texttospeechpb.AudioEncoding_MP3,
			SpeakingRate:  global.VoiceSpeed,
		},
	}

	log.Println("Making the SynthesizeSpeech API call...")
	resp, err := client.SynthesizeSpeech(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to synthesize speech: %v", err)
	}
	if resp == nil {
		return "", fmt.Errorf("received nil response from SynthesizeSpeech")
	}

	// Write the audio content to the file
	if err := os.WriteFile(fullPath, resp.AudioContent, 0644); err != nil {
		return "", fmt.Errorf("failed to write audio content to file: %v", err)
	}

	log.Printf("Audio content written to file: %s\n", fullPath)
	return fullPath, nil
}
