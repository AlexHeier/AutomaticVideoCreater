package editVideo

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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

// Utility function to convert duration to float64 seconds
func durationToSeconds(d *durationpb.Duration) float64 {
	return float64(d.Seconds) + float64(d.Nanos)*1e-9
}

// splitTextIntoLinesWithTimings returns sentences with their timings
func splitTextIntoLinesWithTimings(text string, maxLineLength int, wordTimings []*speechpb.WordInfo) ([]string, []string, error) {
	var lines []string
	var timingStrings []string

	words := strings.Fields(text)
	currentLine := ""
	var lineTimings []*speechpb.WordInfo

	for index, word := range words {
		if index >= len(wordTimings) {
			break
		}

		if len(currentLine)+len(word)+1 > maxLineLength {
			if len(currentLine) > 0 {
				lines = append(lines, currentLine)
				start := durationToSeconds(lineTimings[0].StartTime)
				end := durationToSeconds(lineTimings[len(lineTimings)-1].EndTime)
				timingStrings = append(timingStrings, fmt.Sprintf("between(t,%f,%f)", start, end))
			}
			currentLine = word
			lineTimings = []*speechpb.WordInfo{wordTimings[index]}
		} else {
			if len(currentLine) > 0 {
				currentLine += " "
			}
			currentLine += word
			lineTimings = append(lineTimings, wordTimings[index])
		}
	}

	if len(currentLine) > 0 && len(lineTimings) > 0 {
		lines = append(lines, currentLine)
		start := durationToSeconds(lineTimings[0].StartTime)
		end := durationToSeconds(lineTimings[len(lineTimings)-1].EndTime)
		timingStrings = append(timingStrings, fmt.Sprintf("between(t,%f,%f)", start, end))
	}

	return lines, timingStrings, nil
}

func EditVideo(inputVideoPath string, inputAudioPath string, wordTimings []*speechpb.WordInfo, thema string, content string, author string) (string, error) {
	thema = capitalizeFirstLetter(thema)
	title := fmt.Sprintf("A Quote of %s", thema)
	authorText := fmt.Sprintf("- %s", author)

	// Escape text for FFmpeg
	escapedTitle := escapeText(title)
	contentLines, timingStrings, err := splitTextIntoLinesWithTimings(content, global.CharactersPerLine, wordTimings)
	if err != nil {
		return "", err
	}
	escapedAuthorText := escapeText(authorText)

	// Specify the path to the font file
	fontPath := "fonts/Fredoka-VariableFont_wdth,wght.ttf"

	// Determine the output video path
	outputDir := "edited-videos"
	outputFilename := findNextAvailableFilename(outputDir, fmt.Sprintf("aQuoteOf%s", thema), ".mp4")

	// Build drawtext filters for the content lines
	var drawtextFilters []string
	drawtextFilters = append(drawtextFilters, fmt.Sprintf(
		"drawtext=fontfile='%s':text='%s':x=(w-text_w)/2:y=50:fontsize=110:fontcolor=white:box=1:boxcolor=black@0.5:boxborderw=10",
		fontPath, escapedTitle))
	yPos := 960 // Center vertically for 1080p video
	for i, line := range contentLines {
		if i >= len(timingStrings) {
			return "", fmt.Errorf("timingStrings index %d out of range with length %d", i, len(timingStrings))
		}
		drawtextFilters = append(drawtextFilters, fmt.Sprintf(
			"drawtext=fontfile='%s':text='%s':x=(w-text_w)/2:y=%d:fontsize=100:fontcolor=white:box=1:boxcolor=black@0.5:boxborderw=10:enable='%s'",
			fontPath, escapeText(line), yPos, timingStrings[i]))
	}
	drawtextFilters = append(drawtextFilters, fmt.Sprintf(
		"drawtext=fontfile='%s':text='%s':x=(w-text_w)/2:y=h-th-50:fontsize=100:fontcolor=white:box=1:boxcolor=black@0.5:boxborderw=10",
		fontPath, escapedAuthorText))

	// Combine all drawtext filters
	filterComplex := fmt.Sprintf(
		"[0:v]scale=1080:1920:force_original_aspect_ratio=increase,crop=1080:1920,%s[v];[1:a]volume=2[a]",
		strings.Join(drawtextFilters, ","))

	// FFmpeg command for creating the video with text overlays and adding audio
	cmd := exec.Command("ffmpeg",
		"-i", inputVideoPath,
		"-i", inputAudioPath,
		"-filter_complex", filterComplex,
		"-map", "[v]",
		"-map", "[a]",
		"-c:v", "libx264",
		"-c:a", "aac",
		"-shortest",
		"-y", outputFilename)

	// Run FFmpeg command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("FFmpeg command failed: %v, output: %s", err, string(output))
	}

	log.Printf("Done creating video %s", outputFilename)
	return outputFilename, nil
}
