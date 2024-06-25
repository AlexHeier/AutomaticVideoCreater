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
// filePath is the path to the video file.
// title is the title of the video.
// description is the text description of the video.
func UploadVideoTikTok(filePath, title, description string) error {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Prepare a form that you will submit to TikTok.
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	// Add the video file to the form
	fw, err := w.CreateFormFile("video", filePath)
	if err != nil {
		return err
	}
	if _, err = io.Copy(fw, file); err != nil {
		return err
	}

	// Add the title to the form
	if err := w.WriteField("title", title); err != nil {
		return err
	}

	// Add the description to the form
	if err := w.WriteField("description", description); err != nil {
		return err
	}

	// Add a field to make the video public
	if err := w.WriteField("public", "true"); err != nil {
		return err
	}

	w.Close()

	// Retrieve the API key from an environment variable
	apiKey := os.Getenv("TIKTOK_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("API key not found in the environment variables")
	}

	// Create a request to TikTok's video upload endpoint
	req, err := http.NewRequest("POST", "https://api.tiktok.com/upload/video", &b)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	// Perform the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Process the response (you might want to handle this differently based on your application's needs)
	return nil
}