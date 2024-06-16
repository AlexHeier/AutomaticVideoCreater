package voice

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"videoCreater/global"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	texttospeechpb "cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"google.golang.org/api/option"

	speech "cloud.google.com/go/speech/apiv1"
	speechpb "google.golang.org/genproto/googleapis/cloud/speech/v1"
)

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

	i := 1
	for ; occupied[i]; i++ {
	}
	return filepath.Join(dir, "voice"+strconv.Itoa(i)+".mp3"), nil
}

func ConvertAudioFile(inputPath, outputPath string, channels, sampleRate int) error {
	if inputPath == "" || outputPath == "" {
		return fmt.Errorf("invalid input or output path")
	}

	cmd := exec.Command("ffmpeg", "-i", inputPath, "-ac", fmt.Sprintf("%d", channels), "-ar", fmt.Sprintf("%d", sampleRate), outputPath)
	if _, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ffmpeg error: %v", err)
	}
	return nil
}

func ConvertTextToSpeech(text string) (string, error) {
	ctx := context.Background()

	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("GOOGLE_API_KEY environment variable is not set")
	}

	client, err := texttospeech.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return "", fmt.Errorf("failed to create client: %v", err)
	}
	defer client.Close()

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
			Pitch:         global.Pitch,
		},
	}

	resp, err := client.SynthesizeSpeech(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to synthesize speech: %v", err)
	}
	if resp == nil {
		return "", fmt.Errorf("received nil response from SynthesizeSpeech")
	}

	if err := os.WriteFile(fullPath, resp.AudioContent, 0644); err != nil {
		return "", fmt.Errorf("failed to write audio content to file: %v", err)
	}
	return fullPath, nil
}

func ExtractWordTimings(audioFilePath string) ([]*speechpb.WordInfo, error) {
	ctx := context.Background()

	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GOOGLE_API_KEY environment variable is not set")
	}

	client, err := speech.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create speech client: %v", err)
	}
	defer client.Close()

	convertedFilePath := filepath.Join(filepath.Dir(audioFilePath), "converted_audio.wav")

	if err := ConvertAudioFile(audioFilePath, convertedFilePath, 1, 16000); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(convertedFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio file: %v", err)
	}

	req := &speechpb.LongRunningRecognizeRequest{
		Config: &speechpb.RecognitionConfig{
			Encoding:              speechpb.RecognitionConfig_LINEAR16,
			SampleRateHertz:       16000,
			LanguageCode:          "en-US",
			EnableWordTimeOffsets: true,
			AudioChannelCount:     1,
		},
		Audio: &speechpb.RecognitionAudio{
			AudioSource: &speechpb.RecognitionAudio_Content{Content: data},
		},
	}

	op, err := client.LongRunningRecognize(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to initiate speech recognition: %v", err)
	}
	resp, err := op.Wait(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to complete speech recognition: %v", err)
	}

	var wordInfos []*speechpb.WordInfo
	for _, result := range resp.Results {
		for _, alt := range result.Alternatives {
			wordInfos = append(wordInfos, alt.Words...)
		}
	}
	return wordInfos, nil
}
