package reddit

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

const filename string = "createRedditVideo/last_post_data.json" // File where the last post ID and content are stored
const userAgent string = "windows:videoCreater:v_test (by /u/Heier420)"

type RedditAPIResponse struct {
	Data struct {
		Children []struct {
			Data RedditPost
		}
	}
}

type RedditPost struct {
	Title   string `json:"title"`
	ID      string `json:"id"`
	Content string `json:"selftext"` // This field holds the main body text of the post
}

func getLatestPost(subreddit string) (*RedditPost, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("https://www.reddit.com/r/%s/new.json?limit=1", subreddit), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", userAgent)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result RedditAPIResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if len(result.Data.Children) > 0 {
		return &result.Data.Children[0].Data, nil
	}
	return nil, fmt.Errorf("no posts found")
}

func savePostData(post *RedditPost) error {
	data, err := json.Marshal(post)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

func readPostData() (*RedditPost, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var post RedditPost
	if err := json.Unmarshal(data, &post); err != nil {
		return nil, err
	}
	return &post, nil
}

func GetRedditPost(subreddit string) (*RedditPost, error) {
	// Check if the file exists, create if it does not
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// Create an empty file if it does not exist
		if err := os.WriteFile(filename, []byte("{}"), 0644); err != nil {
			return nil, fmt.Errorf("failed to create initial data file: %v", err)
		}
	}

	lastPost, err := readPostData()
	if err != nil {
		return nil, fmt.Errorf("error reading last post data: %v", err)
	}

	post, err := getLatestPost(subreddit)
	if err != nil {
		return nil, fmt.Errorf("error fetching post: %v", err)
	}

	// Check if the fetched post is new compared to the last saved post
	if lastPost == nil || post.ID != lastPost.ID {
		if err := savePostData(post); err != nil {
			log.Printf("Error saving last post data: %v", err)
		}
		return post, nil // Return the new post
	}

	// Return nil if there is no new post
	return nil, fmt.Errorf("no new posts")
}
