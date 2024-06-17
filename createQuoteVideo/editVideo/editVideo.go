package editVideo

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	"videoCreater/global"

	speechpb "google.golang.org/genproto/googleapis/cloud/speech/v1"
	"google.golang.org/protobuf/types/known/durationpb"
)

// escapeText escapes the text to be used in ffmpeg drawtext filter
func escapeText(text string) string {
	text = strings.ReplaceAll(text, ":", "\\:")
	text = strings.ReplaceAll(text, "'", "'\\''")
	text = strings.ReplaceAll(text, "\n", "")
	return text
}

// capitalizeFirstLetter capitalizes the first letter of the given string
func capitalizeFirstLetter(s string) string {
	for i, v := range s {
		return string(unicode.ToUpper(v)) + s[i+1:]
	}
	return ""
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
func splitTextIntoWordsWithTimings(wordTimings []*speechpb.WordInfo) ([]string, []string) {
	var words []string
	var timingStrings []string

	for i := 0; i < len(wordTimings); i++ {
		word := wordTimings[i].Word
		startTime := durationToSeconds(wordTimings[i].StartTime)
		var endTime float64

		if i < len(wordTimings)-1 {
			endTime = durationToSeconds(wordTimings[i+1].StartTime)
		} else {
			endTime = durationToSeconds(wordTimings[i].EndTime)
		}

		words = append(words, word)
		timingStrings = append(timingStrings, fmt.Sprintf("between(t,%f,%f)", startTime, endTime))
	}

	return words, timingStrings
}

// Utility function to convert duration to float64 seconds
func durationToSeconds(d *durationpb.Duration) float64 {
	return float64(d.Seconds) + float64(d.Nanos)*1e-9
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

func EditVideo(inputVideoPath string, inputAudioPath string, wordTimings []*speechpb.WordInfo, thema string, author string) (string, error) {
	thema = capitalizeFirstLetter(thema)
	title := fmt.Sprintf("A Quote of %s", thema)
	authorText := fmt.Sprintf("- %s", abbreviateAuthorName(author))

	// Get audio duration using ffprobe
	audioDuration, err := getVideoDuration(inputAudioPath)
	if err != nil {
		return "", err
	}

	// Get video duration using ffprobe
	videoDuration, err := getVideoDuration(inputVideoPath)
	if err != nil {
		return "", err
	}

	// Calculate the number of times to loop the video
	loopCount := int(audioDuration / videoDuration)

	// Escape text for FFmpeg
	escapedTitle := escapeText(title)
	words, timingStrings := splitTextIntoWordsWithTimings(wordTimings)
	escapedAuthorText := escapeText(authorText)

	// Specify the path to the font file
	fontPath := "fonts/PermanentMarker-Regular.ttf"

	// Determine the output video path
	outputDir := "edited-videos"
	outputFilename := findNextAvailableFilename(outputDir, fmt.Sprintf("aQuoteOf%s", thema), ".mp4")

	// Build drawtext filters for the content lines
	var drawtextFilters []string
	// Title near the top of the screen
	drawtextFilters = append(drawtextFilters, fmt.Sprintf(
		"drawtext=fontfile='%s':text='%s':x=(w-text_w)/2:y=50:fontsize=110:fontcolor=white:borderw=%d:bordercolor=black",
		fontPath, escapedTitle, global.BorderThickness))
	// Words centered in the middle of the screen
	for i, word := range words {
		drawtextFilters = append(drawtextFilters, fmt.Sprintf(
			"drawtext=fontfile='%s':text='%s':x=(w-text_w)/2:y=(h/2-text_h/2):fontsize=100:fontcolor=white:borderw=%d:bordercolor=black:enable='%s'",
			fontPath, escapeText(word), global.BorderThickness, timingStrings[i]))
	}
	// Author text at the bottom of the screen
	drawtextFilters = append(drawtextFilters, fmt.Sprintf(
		"drawtext=fontfile='%s':text='%s':x=(w-text_w)/2:y=h-th-50:fontsize=100:fontcolor=white:borderw=%d:bordercolor=black",
		fontPath, escapedAuthorText, global.BorderThickness))

	// Combine all drawtext filters
	filterComplex := fmt.Sprintf(
		"[0:v]scale=1080:1920:force_original_aspect_ratio=increase,crop=1080:1920,%s[v];[1:a]volume=2[a]",
		strings.Join(drawtextFilters, ","))

	// FFmpeg command for creating the video with text overlays and adding audio, and looping the video if necessary
	cmdArgs := []string{
		"-stream_loop", strconv.Itoa(loopCount), // Loop the video input
		"-i", inputVideoPath,
		"-i", inputAudioPath,
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

	fmt.Printf("Done creating video %s", outputFilename)
	return outputFilename, nil
}
