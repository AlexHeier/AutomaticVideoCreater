package quote

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
	"videoCreater/global"
)

// Quote struct to match the JSON structure
type Quote struct {
	Body   string `json:"body"`
	Author string `json:"author"`
}

// QuotesResponse struct to match the response structure
type QuotesResponse struct {
	Quotes []Quote `json:"quotes"`
	Page   int     `json:"page"`
	Total  int     `json:"total"`
}

// Function that fetches a quote and returns the content and author
func FetchQuote(thema string) (string, string, error) {
	// Get the API key from environment variables
	apiKey := os.Getenv("FAVQS_API_KEY")
	if apiKey == "" {
		return "", "", fmt.Errorf("FAVQS_API_KEY environment variable not set")
	}
	log.Println(thema)

	// Create a new random generator
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for {
		// Generate a random page number between 1 and 10
		randomPage := rng.Intn(10) + 1

		// URL from which to fetch the quote
		pageURL := fmt.Sprintf("https://favqs.com/api/quotes/?filter=%s&type=tag&page=%d", thema, randomPage)

		// Create a new request
		req, err := http.NewRequest("GET", pageURL, nil)
		if err != nil {
			log.Printf("Error creating request: %v. Retrying...", err)
			time.Sleep(3 * time.Second)
			continue
		}

		req.Header.Set("Authorization", "Token token="+apiKey)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Error making request: %v. Retrying...", err)
			time.Sleep(3 * time.Second)
			continue
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Error reading response: %v. Retrying...", err)
			time.Sleep(3 * time.Second)
			continue
		}

		var quotesResponse QuotesResponse
		err = json.Unmarshal(body, &quotesResponse)
		if err != nil {
			fmt.Printf("Received JSON: %s\n", body)
			log.Printf("Error parsing JSON: %v. Retrying...", err)
			time.Sleep(3 * time.Second)
			continue
		}

		if len(quotesResponse.Quotes) == 0 {
			log.Println("No quotes found. Retrying...")
			time.Sleep(3 * time.Second)
			continue
		}

		// Select a random quote from the fetched quotes
		randomQuote := quotesResponse.Quotes[rng.Intn(len(quotesResponse.Quotes))]

		if randomQuote.Body == "No quotes found" {
			// Remove the current thema from the global list
			for i, t := range global.Themas {
				if t == thema {
					global.Themas = append(global.Themas[:i], global.Themas[i+1:]...)
					// Return an error stating that no quotes were found for this thema
					return "", "", fmt.Errorf("no quotes found for the thema: %s", thema)
				}
			}
		}

		// Check if the content length is less than or equal to 200 characters
		if len(randomQuote.Body) <= 200 {
			return randomQuote.Body, randomQuote.Author, nil
		} else {
			log.Println("Selected quote does not meet length criteria. Retrying...")
			time.Sleep(3 * time.Second)
		}
	}
}
