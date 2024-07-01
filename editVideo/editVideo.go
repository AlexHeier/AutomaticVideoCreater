package editVideo

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	voice "videoCreater/voice"
)

// escapeText escapes the text to be used in ffmpeg drawtext filter
func escapeText(text string) string {
	text = strings.ReplaceAll(text, ":", "\\:")
	text = strings.ReplaceAll(text, "'", "'\\''")
	text = strings.ReplaceAll(text, "\n", "")
	return text
}

// findNextAvailableFilename finds the next available filename with the given prefix
func findNextAvailableFilename(dir, prefix, extension string) string {
	for i := 1; ; i++ {
		filename := fmt.Sprintf("%s_%d%s", prefix, i, extension)
		fullPath := filepath.Join(dir, filename)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			return fullPath
		}
	}
}

// abbreviateAuthorName shortens the author's name if it is longer than 20 characters
func abbreviateAuthorName(author string) string {
	if len(author) > 20 {
		words := strings.Fields(author)
		if len(words) > 1 {
			return fmt.Sprintf("%s %s", words[0], words[len(words)-1])
		}
	}
	return author
}

// splitTextIntoWordsWithTimings splits the text into words and calculates the timings for each word.
func splitTextIntoWordsWithTimingsWithAuthor(wordTimings []voice.WordInfo, author string, audioDuration float64) ([]string, []string) {
	var words []string
	var timingStrings []string
	var endTime float64

	for i := 0; i < len(wordTimings); i++ {
		word := wordTimings[i].Word
		startTime := wordTimings[i].StartTime

		if i < len(wordTimings)-1 {
			endTime = wordTimings[i+1].StartTime
		} else {
			endTime = wordTimings[i].EndTime
		}

		words = append(words, word)
		timingStrings = append(timingStrings, fmt.Sprintf("between(t,%f,%f)", startTime, endTime))
	}

	// Add the author's name with a timing string starting at the end of the last word and lasting for 1 second
	words = append(words, author)
	timingStrings = append(timingStrings, fmt.Sprintf("between(t,%f,%f)", endTime, audioDuration))

	return words, timingStrings
}

func splitTextIntoWordsWithTimings(wordTimings []voice.WordInfo) ([]string, []string, float64) {
	var words []string
	var timingStrings []string
	var endTime float64

	for i := 0; i < len(wordTimings); i++ {
		word := wordTimings[i].Word
		startTime := wordTimings[i].StartTime

		if i < len(wordTimings)-1 {
			endTime = wordTimings[i+1].StartTime
		} else {
			endTime = wordTimings[i].EndTime
		}

		words = append(words, word)
		timingStrings = append(timingStrings, fmt.Sprintf("between(t,%f,%f)", startTime, endTime))
	}

	return words, timingStrings, endTime
}

func getVideoDuration(videoPath string) (float64, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", videoPath)
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to get video duration: %v", err)
	}
	videoDuration, err := strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse video duration: %v", err)
	}
	return videoDuration, nil
}

func removeSpaces(input string) string {
	return strings.ReplaceAll(input, " ", "")
}

func splitTitleIntoLines(title string, maxLength int) []string {
	var lines []string
	words := strings.Fields(title)
	var currentLine string

	for _, word := range words {
		if len(currentLine)+len(word)+1 > maxLength {
			lines = append(lines, currentLine)
			currentLine = word + " "
		} else {
			currentLine += word + " "
		}
	}

	// Add the last line if it's not empty
	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	// If there are more than 4 lines, truncate and add "..."
	if len(lines) > 4 {
		lines = lines[:4]
		lines[3] = lines[3] + "..."
	}

	return lines
}

// DownloadImage downloads an image from the given URL and saves it to the specified path
func DownloadImage(url, relativePath string) error {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download image: %v", err)
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned non-200 status: %d %s", resp.StatusCode, resp.Status)
	}

	path, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working dir: %v", err)
	}

	fullPath := filepath.Join(path, relativePath)
	log.Println(fullPath)

	// Create the file
	out, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write to file: %v", err)
	}

	// Set read and write permissions for owner, read for group and others
	if err := os.Chmod(fullPath, 0644); err != nil {
		return fmt.Errorf("failed to set file permissions: %v", err)
	}

	return nil
}
