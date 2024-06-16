package editVideo

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"
)

// escapeText escapes the text to be used in ffmpeg drawtext filter
func escapeText(text string) string {
	text = strings.ReplaceAll(text, ":", "\\:")
	text = strings.ReplaceAll(text, "'", "'\\''")
	text = strings.ReplaceAll(text, "\n", "")
	return text
}

// splitTextIntoLines splits the text into lines of a specified maximum length
func splitTextIntoLines(text string, maxLineLength int) []string {
	var lines []string
	escape := escapeText(text)
	words := strings.Fields(escape)
	var currentLine string
	for _, word := range words {
		if len(currentLine)+len(word)+1 > maxLineLength {
			lines = append(lines, currentLine)
			currentLine = word
		} else {
			if len(currentLine) > 0 {
				currentLine += " "
			}
			currentLine += word
		}
	}
	if len(currentLine) > 0 {
		lines = append(lines, currentLine)
	}
	return lines
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

// EditVideo adds a title, quote, and author to the video, adds audio, and saves it to a new file
func EditVideo(inputVideoPath, inputAudioPath, thema, content, author string) (string, error) {
	thema = capitalizeFirstLetter(thema)
	title := fmt.Sprintf("A Quote of %s", thema)
	authorText := fmt.Sprintf("- %s", author)

	// Escape text for ffmpeg
	escapedTitle := escapeText(title)
	contentLines := splitTextIntoLines(content, 20)
	escapedAuthorText := escapeText(authorText)

	// Specify the path to the Fredoka font file
	fontPath := "fonts/Fredoka-VariableFont_wdth,wght.ttf"

	// Determine the output video path
	outputDir := "edited-videos"
	outputFilename := findNextAvailableFilename(outputDir, fmt.Sprintf("aQuoteOf%s", thema), ".mp4")

	// Build drawtext filters for the content lines
	var drawtextFilters []string
	drawtextFilters = append(drawtextFilters, fmt.Sprintf(
		"drawtext=fontfile='%s':text='%s':x=(w-text_w)/2:y=50:fontsize=110:fontcolor=white:box=1:boxcolor=black@0.5:boxborderw=10",
		fontPath, escapedTitle))
	yPos := 350
	for _, line := range contentLines {
		drawtextFilters = append(drawtextFilters, fmt.Sprintf(
			"drawtext=fontfile='%s':text='%s':x=(w-text_w)/2:y=%d:fontsize=100:fontcolor=white:box=1:boxcolor=black@0.5:boxborderw=10",
			fontPath, escapeText(line), yPos))
		yPos += 110 // Increase y position for each line
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

	// Capture standard output and error
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to run ffmpeg command: %v, output: %s", err, string(output))
	}

	return outputFilename, nil
}
