package editVideo

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"videoCreater/global"
	voice "videoCreater/voice"
)

// EditVideoTikTok creates TikTok-style videos with text overlays from input video and audio files.
func EditVideoTikTok(inputVideoPath string, inputAudioPaths []string, wordTimings [][]voice.WordInfo, title string) ([]string, error) {
	const fontSize = 110

	titleFontSize := 110
	lineHeight := float64(titleFontSize) * 1.2

	charsPerLine := 15 // Initial estimate

	var outputFilenames []string
	var elapsedTime float64

	for i := range inputAudioPaths {

		// Escape text for FFmpeg
		words, timingStrings, endTime := splitTextIntoWordsWithTimings(wordTimings[i])

		// Specify the path to the font file
		fontPath := "fonts/PermanentMarker-Regular.ttf"

		// TikTok logo image
		tikTokLogo := "logos/tiktok_logo.png"

		// Determine the output video path
		outputDir := "edited-videos"
		partSuffix := ""
		if len(inputAudioPaths) > 1 {
			partSuffix = fmt.Sprintf("Part%dof%d", i+1, len(inputAudioPaths))
		}

		shortTitle := title
		if len(title) > 25 {
			shortTitle = title[:25]
		}
		outputFilename := findNextAvailableFilename(outputDir, removeSpaces(shortTitle+partSuffix), ".mp4")

		titleLines := splitTitleIntoLines(title, charsPerLine)

		// Build drawtext filters for the content lines
		var drawtextFilters []string
		// Title near the top of the screen
		for j, line := range titleLines {
			drawtextFilters = append(drawtextFilters, fmt.Sprintf(
				"drawtext=fontfile='%s':text='%s':x=(w-text_w)/2:y=50+(%d*%.0f):fontsize=%d:fontcolor=white:borderw=%d:bordercolor=black",
				fontPath, escapeText(line), j, lineHeight, titleFontSize, global.BorderThickness))
		}
		// Part text under the title in the middle
		if partSuffix != "" {
			drawtextFilters = append(drawtextFilters, fmt.Sprintf(
				"drawtext=fontfile='%s':text='%s':x=(w-text_w)/2:y=h-th-200:fontsize=80:fontcolor=white:borderw=%d:bordercolor=black",
				fontPath, escapeText(fmt.Sprintf("part %v of %v", i+1, len(inputAudioPaths))), global.BorderThickness))
		}
		// Words centered in the middle of the screen
		for j, word := range words {
			drawtextFilters = append(drawtextFilters, fmt.Sprintf(
				"drawtext=fontfile='%s':text='%s':x=(w-text_w)/2:y=(h/2-text_h/2):fontsize=100:fontcolor=white:borderw=%d:bordercolor=black:enable='%s'",
				fontPath, escapeText(word), global.BorderThickness, timingStrings[j]))
		}
		// Channel name text at the bottom of the screen
		drawtextFilters = append(drawtextFilters, fmt.Sprintf(
			"drawtext=fontfile='%s':text='%s':x=(w-text_w)/2:y=h-th-50:fontsize=%d:fontcolor=white:borderw=%d:bordercolor=black",
			fontPath, global.TikTokChannelName, fontSize, global.BorderThickness))

		// Combine all drawtext filters
		filterComplex := fmt.Sprintf(
			"[0:v]trim=start=%f,setpts=PTS-STARTPTS,scale=1080:1920:force_original_aspect_ratio=increase,crop=1080:1920,%s[v];[1:a]volume=2[a];[2:v]scale=-1:%d[tiktok_logo];[v][tiktok_logo]overlay=x=25:y=main_h-overlay_h-50[v]",
			elapsedTime+60, strings.Join(drawtextFilters, ","), fontSize) // Add 60 seconds to the elapsed time

		// Write filter complex to a temporary file
		filterFile, err := os.CreateTemp("", "ffmpeg-filter-*.txt")
		if err != nil {
			return nil, fmt.Errorf("failed to create temp file: %v", err)
		}
		defer os.Remove(filterFile.Name()) // Clean up the temp file afterwards

		if _, err := filterFile.WriteString(filterComplex); err != nil {
			return nil, fmt.Errorf("failed to write to temp file: %v", err)
		}
		if err := filterFile.Close(); err != nil {
			return nil, fmt.Errorf("failed to close temp file: %v", err)
		}

		// FFmpeg command for creating the video with text overlays and adding audio, and looping the video if necessary
		cmdArgs := []string{
			"-i", inputVideoPath,
			"-i", inputAudioPaths[i],
			"-i", tikTokLogo,
			"-filter_complex_script", filterFile.Name(),
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
			return nil, fmt.Errorf("FFmpeg command failed: %v, output: %s", err, string(output))
		}

		fmt.Printf("Done creating video %s\n", outputFilename)
		outputFilenames = append(outputFilenames, outputFilename)

		// Update the elapsed time
		elapsedTime += endTime
	}

	return outputFilenames, nil
}
