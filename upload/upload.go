package upload

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"videoCreater/global"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

// getClient retrieves the authenticated HTTP client using OAuth2
func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
	tokenFile := "token.json"
	token, err := tokenFromFile(tokenFile)
	if err != nil {
		token = getTokenFromWeb(ctx, config)
		saveToken(tokenFile, token)
	}
	return config.Client(ctx, token)
}

// tokenFromFile retrieves the token from the file
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	token := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(token)
	return token, err
}

// getTokenFromWeb retrieves the token from the web by prompting the user
func getTokenFromWeb(ctx context.Context, config *oauth2.Config) *oauth2.Token {
	state := "state-token"
	authURL := config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the authorization code: \n%v\n", authURL)

	codeChan := make(chan string)
	// Start a local server to handle the redirect
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if query.Get("state") != state {
			http.Error(w, "State does not match", http.StatusBadRequest)
			return
		}
		code := query.Get("code")
		if code == "" {
			http.Error(w, "Code not found", http.StatusBadRequest)
			return
		}
		fmt.Fprintf(w, "Authorization complete, you can close this window.")
		codeChan <- code
	})

	go http.ListenAndServe(":8080", nil)

	authCode := <-codeChan

	token, err := config.Exchange(ctx, authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return token
}

// saveToken saves the token to a file for later use
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.Create(path)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

// sanitizeDescription ensures the description is valid for YouTube
// sanitizeDescription ensures the description is valid for YouTube
func sanitizeDescription(baseDescription string, tags []string) string {
	// Initial formatting and sanitization
	description := strings.TrimSpace(baseDescription)
	description = strings.ReplaceAll(description, "\n", " ") // Replace new lines with spaces
	description = strings.ReplaceAll(description, "\r", "")
	description = strings.ReplaceAll(description, "\t", " ") // Replace tabs with spaces

	// Append tags as hashtags for SEO
	hashtagSection := "\n\n"
	for _, tag := range tags {
		hashtagSection += fmt.Sprintf("#%s ", tag)
	}
	if len(description)+len(hashtagSection) > 5000 {
		description = description[:5000-len(hashtagSection)-3] + "..." // Truncate to fit
	}
	description += hashtagSection

	return description
}

func UploadVideo(videoPath string, thema string) error {
	ctx := context.Background()

	b, err := os.ReadFile("client_secret.json")
	if err != nil {
		return fmt.Errorf("unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, youtube.YoutubeUploadScope)
	if err != nil {
		return fmt.Errorf("unable to parse client secret file to config: %v", err)
	}
	config.RedirectURL = "http://localhost:8080"

	client := getClient(ctx, config)

	service, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("unable to create YouTube service: %v", err)
	}

	// Capitalize the first letter of the theme using the cases package
	titleCase := cases.Title(language.English, cases.NoLower)
	themaCapitalized := titleCase.String(thema)
	title := fmt.Sprintf("A Quote of %s", themaCapitalized)

	description := fmt.Sprintf("A beautiful quote about %s. Leave a Like and Subscribe for more beautiful quotes <3 ", themaCapitalized)

	tags := make([]string, 0, global.TagLimit) // Initialize an empty slice with capacity for tagLimit elements

	for i, t := range global.Themas {
		if i >= global.TagLimit {
			break // Stop the loop if the limit is reached
		}
		tags = append(tags, t) // Append each tag to the slice
	}

	// Sanitize description
	description = sanitizeDescription(description, tags)
	log.Println(description)
	video := &youtube.Video{
		Snippet: &youtube.VideoSnippet{
			Title:       title,
			Description: description,
			Tags:        tags,
			CategoryId:  "22", // People & Blogs category
		},
		Status: &youtube.VideoStatus{PrivacyStatus: "public"},
	}

	call := service.Videos.Insert([]string{"snippet", "status"}, video)

	file, err := os.Open(videoPath)
	if err != nil {
		return fmt.Errorf("error opening video file: %v", err)
	}
	defer file.Close()

	response, err := call.Media(file).Do()
	if err != nil {
		return fmt.Errorf("error uploading video: %v", err)
	}

	fmt.Printf("Upload successful! Video ID: %v\n", response.Id)
	return nil
}
