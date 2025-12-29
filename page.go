package ditch

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	// itch.io hostname
	hostname = "https://itch.io"

	// default parameters when doing API calls
	onsaleParams = "/on-sale?format=json&page"

	// max retries for API calls
	maxRetries = 5

	// max number of items returned by an API call
	maxItems = 36
)

// fetchDataWithRetry returns the response of an HTTP GET request for a given URL.
// It returns the response and a detailed error if any.
func fetchDataWithRetry(url string) (*http.Response, error) {
	for retries := 0; retries < maxRetries; retries++ {
		resp, err := http.Get(url)

		if err != nil {
			return nil, err
		}

		if resp.StatusCode == 429 && retries < maxRetries-1 {
			fmt.Printf("Failed to fetch data from %s. Status code 429. Retrying in 2 seconds...\n", url)
			resp.Body.Close()
			time.Sleep(2 * time.Second)
			continue
		}

		if resp.StatusCode != 200 && resp.StatusCode != 429 {
			resp.Body.Close()
			return nil, fmt.Errorf("Failed to fetch data from %s. Status code %d, ", url, resp.StatusCode)
		}

		return resp, nil
	}
	return nil, fmt.Errorf("Failed to fetch data from %s after %d retries", url, maxRetries)
}

// getJSON returns the content of a page for a given category.
// It returns the JSON as a string and an error if any.
func getJSON(category string, page int) (string, error) {
	url := fmt.Sprintf("%s/%s%s=%d", hostname, category, onsaleParams, page)

	resp, err := fetchDataWithRetry(url)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	return string(body), nil
}

// getSales returns the content of a sales page and an error if any.
func getSales(link string) (string, error) {
	url := fmt.Sprintf("%s%s", hostname, link)
	resp, err := fetchDataWithRetry(url)

	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// getGame returns the content of a game page and an error if any.
func getGame(link string) (string, error) {
	url := link
	if !strings.HasPrefix(link, "http://") && !strings.HasPrefix(link, "https://") {
		url = fmt.Sprintf("%s%s", hostname, link)
	}
	resp, err := fetchDataWithRetry(url)

	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// getContent puts the content of a page in a list for a given category.
// It returns whether it was the last page and an error if any.
func getContent(category string, page int, list *[]Content) (isLastPage bool, err error) {
	json, err := getJSON(category, page)
	if err != nil {
		fmt.Printf(`
		Function: getContent::getJSON
		Context:
		- category: %s
		- page:     %d

		Error: %s
		`, category, page, err)
		return isLastPage, err
	}

	content := Content{}
	err = content.FromJSON(json)
	if err != nil {
		fmt.Printf(`
		Function: getContent::content.FromJSON
		Context:
		- category: %s
		- page:     %d
		- json:	    %s

		Error: %s
		`, category, page, json, err)
		return isLastPage, err
	}

	*list = append(*list, content)

	isLastPage = content.NumItems < maxItems

	return isLastPage, nil
}

// Type that represents a function to get a Content for a specific category.
type GetCategoryContentFn func(int, *[]Content) (bool, error)

// GetGamesContent puts in a list the `games` type content for a given page.
// It returns whether it was the last page and an error if any.
func GetGamesContent(page int, list *[]Content) (isLastPage bool, err error) {
	return getContent("games", page, list)
}
