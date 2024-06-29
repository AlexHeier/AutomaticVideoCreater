package editVideo

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"videoCreater/global"
	voice "videoCreater/voice"
)

func EditVideoYoutube(inputVideoPath string, inputAudioPath string, wordTimings []voice.WordInfo, title string, author string) (string, error) {
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
	words, timingStrings := splitTextIntoWordsWithTimingsWithAuthor(wordTimings, escapedAuthorText, audioDuration)

	// Specify the path to the font file
	fontPath := "fonts/PermanentMarker-Regular.ttf"

	// Download the YouTube logo image
	youtubeLogoURL := "https://upload.wikimedia.org/wikipedia/commons/e/ef/Youtube_logo.png"
	youtubeLogoPath := filepath.Join("logos", "youtube_logo.png")
	err = DownloadImage(youtubeLogoURL, youtubeLogoPath)
	if err != nil {
		return "", fmt.Errorf("failed to download YouTube logo: %v", err)
	}

	// Ensure the image is deleted after the function completes
	defer os.Remove(youtubeLogoPath)

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
		fontPath, global.YoutubeChannelName, fontSize, global.BorderThickness))

	// Combine all drawtext filters
	filterComplex := fmt.Sprintf(
		"[0:v]scale=1080:1920:force_original_aspect_ratio=increase,crop=1080:1920,%s[v];[1:a]volume=2[a];[2:v]scale=-1:%d[youtube_logo];[v][youtube_logo]overlay=x=25:y=main_h-overlay_h-50[v]",
		strings.Join(drawtextFilters, ","), fontSize)

	// FFmpeg command for creating the video with text overlays and adding audio, and looping the video if necessary
	cmdArgs := []string{
		"-stream_loop", "-1", // Infinite loop for video
		"-i", inputVideoPath,
		"-i", inputAudioPath,
		"-i", youtubeLogoPath, // Add the YouTube logo image
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
