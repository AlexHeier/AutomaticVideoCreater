package editVideo

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"videoCreater/global"
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
		filename := fmt.Sprintf("%s%d%s", prefix, i, extension)
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
func splitTextIntoWordsWithTimings(wordTimings []voice.WordInfo, author string, audioDuration float64) ([]string, []string) {
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
	if currentLine != "" {
		lines = append(lines, currentLine)
	}
	return lines
}

func EditVideo(inputVideoPath string, inputAudioPath string, wordTimings []voice.WordInfo, title string, author string) (string, error) {
	authorText := fmt.Sprintf("- %s", abbreviateAuthorName(author))
	fontSize := 100         // Set the font size for the author text
	lineHeight := 110 * 1.2 // Set the line height for the title text

	// Get audio duration using ffprobe
	audioDuration, err := getVideoDuration(inputAudioPath)
	if err != nil {
		return "", err
	}

	// Escape text for FFmpeg
	escapedAuthorText := escapeText(authorText)
	words, timingStrings := splitTextIntoWordsWithTimings(wordTimings, escapedAuthorText, audioDuration)

	// Specify the path to the font file
	fontPath := "fonts/PermanentMarker-Regular.ttf"

	// Determine the output video path
	outputDir := "edited-videos"
	outputFilename := findNextAvailableFilename(outputDir, removeSpaces(title), ".mp4")

	titleLines := splitTitleIntoLines(title, 15)
	// Build drawtext filters for the content lines
	var drawtextFilters []string
	// Title near the top of the screen
	for i, line := range titleLines {
		drawtextFilters = append(drawtextFilters, fmt.Sprintf(
			"drawtext=fontfile='%s':text='%s':x=(w-text_w)/2:y=50+(%d*%.0f):fontsize=110:fontcolor=white:borderw=%d:bordercolor=black",
			fontPath, escapeText(line), i, lineHeight, global.BorderThickness))
	}
	// Words centered in the middle of the screen
	for i, word := range words {
		drawtextFilters = append(drawtextFilters, fmt.Sprintf(
			"drawtext=fontfile='%s':text='%s':x=(w-text_w)/2:y=(h/2-text_h/2):fontsize=100:fontcolor=white:borderw=%d:bordercolor=black:enable='%s'",
			fontPath, escapeText(word), global.BorderThickness, timingStrings[i]))
	}
	// Channel name text at the bottom of the screen
	drawtextFilters = append(drawtextFilters, fmt.Sprintf(
		"drawtext=fontfile='%s':text='%s':x=(w-text_w)/2:y=h-th-50:fontsize=%d:fontcolor=white:borderw=%d:bordercolor=black",
		fontPath, global.ChannelName, fontSize, global.BorderThickness))

	// Combine all drawtext filters
	filterComplex := fmt.Sprintf(
		"[0:v]scale=1080:1920:force_original_aspect_ratio=increase,crop=1080:1920,%s[v];[1:a]volume=2[a];[2:v]scale=-1:%d[youtube_logo];[v][youtube_logo]overlay=x=25:y=main_h-overlay_h-50[v]",
		strings.Join(drawtextFilters, ","), fontSize)

	// FFmpeg command for creating the video with text overlays and adding audio, and looping the video if necessary
	cmdArgs := []string{
		"-stream_loop", "-1", // Infinite loop for video
		"-i", inputVideoPath,
		"-i", inputAudioPath,
		"-i", "youtube_logo.png", // Add the YouTube logo image
		"-filter_complex", filterComplex,
		"-map", "[v]",
		"-map", "[a]",
		"-c:v", "libx264",
		"-c:a", "aac",
		"-shortest",
		"-y", outputFilename,
	}

	// FFmpeg command for creating the video with text overlays and adding audio
	cmd := exec.Command("ffmpeg", cmdArgs...)

	// Run FFmpeg command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("FFmpeg command failed: %v, output: %s", err, string(output))
	}

	fmt.Printf("Done creating video %s\n", outputFilename)
	return outputFilename, nil
}
