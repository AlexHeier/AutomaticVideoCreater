package upload

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

// UploadVideoTikTok uploads a video file to TikTok along with a title, description, and makes it public.
func UploadVideoTikTok(filePath, title, description string) error {
	// Open the video file
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Prepare a multipart form to send via POST
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	// Add the video file part
	fw, err := w.CreateFormFile("video", filePath)
	if err != nil {
		return err
	}
	if _, err = io.Copy(fw, file); err != nil {
		return err
	}

	// Add title and description to the form
	if err := w.WriteField("title", title); err != nil {
		return err
	}
	if err := w.WriteField("description", description); err != nil {
		return err
	}

	// Set the video to be public
	if err := w.WriteField("public", "true"); err != nil {
		return err
	}

	w.Close()

	// Retrieve the API key from an environment variable
	apiKey := os.Getenv("TIKTOK_CLIENT_KEY")
	if apiKey == "" {
		return fmt.Errorf("API key not found in the environment variables")
	}

	// Create the POST request
	req, err := http.NewRequest("POST", "https://api.tiktok.com/upload/video", &b)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	return nil
}
