package voice

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"videoCreater/global"
)

// UnrealSpeechResponse is the struct to unmarshal the response from UnrealSpeech's API
type UnrealSpeechResponse struct {
	OutputUri     string `json:"OutputUri"`
	TimestampsUri string `json:"TimestampsUri"`
}

// WordInfo contains the timing information for a word
type WordInfo struct {
	StartTime float64 `json:"start"` // The start time of the word in the audio
	EndTime   float64 `json:"end"`   // The end time of the word in the audio
	Word      string  `json:"word"`  // The word text
}

// findNextAvailableFilename finds the next available filename in the given directory with the specified extension
func findNextAvailableFilename(dir string, extension string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}

	occupied := make(map[int]bool)
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, "voice") && strings.HasSuffix(name, extension) {
			numStr := strings.TrimSuffix(strings.TrimPrefix(name, "voice"), extension)
			num, err := strconv.Atoi(numStr)
			if err == nil {
				occupied[num] = true
			}
		}
	}

	i := 1
	for ; occupied[i]; i++ {
	}
	return filepath.Join(dir, fmt.Sprintf("voice%d%s", i, extension)), nil
}

// downloadFile downloads a file from the given URL and writes it to the given path
func downloadFile(url string, path string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download file from URL %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download file from URL %s: status code %d", url, resp.StatusCode)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %v", path, err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %v", path, err)
	}

	return nil
}

// ConvertTextToSpeech sends text to UnrealSpeech API and returns the path to the saved MP3 file and the timing information of words
func ConvertTextToSpeech(text string) ([]string, [][]WordInfo, error) {
	chunks := assembleChunks(text)
	var paths []string
	var allWordInfos [][]WordInfo

	for _, chunk := range chunks {
		path, wordInfos, err := processTextChunk(chunk)
		if err != nil {
			return nil, nil, err
		}
		paths = append(paths, path)
		allWordInfos = append(allWordInfos, wordInfos)
	}

	return paths, allWordInfos, nil
}

// assembleChunks divides text into chunks that do not exceed the maximum size defined in global.MaxVoiceCharacters, respecting sentence boundaries.
func assembleChunks(text string) []string {
	var chunks []string
	var currentChunk strings.Builder
	sentences := strings.FieldsFunc(text, func(r rune) bool {
		return r == '.' || r == '?' || r == '!'
	})

	currentLength := 0
	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence) + " " // Add the punctuation back with a space for natural reading
		sentenceLength := len(sentence)

		if currentLength+sentenceLength > global.MaxVoiceCharacters {
			if currentChunk.Len() > 0 {
				chunks = append(chunks, currentChunk.String())
				currentChunk.Reset()
				currentLength = 0
			}
		}

		currentChunk.WriteString(sentence)
		currentLength += sentenceLength
	}

	// Ensure the last chunk is added
	if currentChunk.Len() > 0 {
		chunks = append(chunks, currentChunk.String())
	}

	return chunks
}

// processTextChunk handles the interaction with the UnrealSpeech API for a single text chunk
func processTextChunk(text string) (string, []WordInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	apiKey := os.Getenv("UNREAL_SPEECH_API_KEY")
	if apiKey == "" {
		log.Println("API key is not set")
		return "", nil, fmt.Errorf("UNREAL_SPEECH_API_KEY environment variable is not set")
	}

	reqBody := map[string]interface{}{
		"Text":          text,
		"VoiceId":       global.VoiceID,
		"Bitrate":       global.Bitrate,
		"Speed":         global.VoiceSpeed,
		"Pitch":         global.VoicePitch,
		"TimestampType": "word",
	}
	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	url := "https://api.v6.unrealspeech.com/speech"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBytes))
	if err != nil {
		return "", nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("failed to send request to UnrealSpeech API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("API returned non-OK status: %s, body: %s", resp.Status, string(body))
		return "", nil, fmt.Errorf("UnrealSpeech API returned non-OK status: %s, body: %s", resp.Status, string(body))
	}

	var apiResponse UnrealSpeechResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return "", nil, fmt.Errorf("failed to decode response: %v", err)
	}

	dir := "text-to-speeched"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.Mkdir(dir, 0755); err != nil {
			return "", nil, fmt.Errorf("failed to create directory: %v", err)
		}
	}

	path, err := findNextAvailableFilename(dir, ".mp3")
	if err != nil {
		return "", nil, err
	}

	// Download the MP3 file from OutputUri
	err = downloadFile(apiResponse.OutputUri, path)
	if err != nil {
		return "", nil, fmt.Errorf("failed to download MP3 file: %v", err)
	}

	// Download the JSON file from TimestampsUri
	resp, err = http.Get(apiResponse.TimestampsUri)
	if err != nil {
		return "", nil, fmt.Errorf("failed to download JSON from URL %s: %v", apiResponse.TimestampsUri, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("failed to download JSON from URL %s: status code %d", apiResponse.TimestampsUri, resp.StatusCode)
	}

	var wordInfos []WordInfo
	if err := json.NewDecoder(resp.Body).Decode(&wordInfos); err != nil {
		return "", nil, fmt.Errorf("failed to decode JSON response: %v", err)
	}

	return path, wordInfos, nil
}
