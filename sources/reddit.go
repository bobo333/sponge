package sources

import (
	"encoding/json"
	"fmt"
	shared "github.com/bobo333/sponge/shared"
	"net/http"
	"time"
)

type redditItem struct {
	Title     string `json:"title"`
	Url       string `json:"url"`
	Permalink string `json:"permalink"`
}

func (r redditItem) standardize() shared.StandardizedItem {
	return shared.StandardizedItem{
		Title:    r.Title,
		Url:      r.Url,
		Comments: fmt.Sprintf("https://reddit.com%s", r.Permalink)}
}

type redditList struct {
	Data struct {
		Children []struct {
			Data redditItem `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

// GetReddit retrieves *numItems* from *subName* subreddit, standardizes them,
// and compiles them into shared.OutputSection.
func GetReddit(subName string, numItems int) (shared.OutputSection, error) {
	redditUsernameEnvName := "REDDIT_USERNAME"
	redditUsername, envVarErr := shared.GetEnvVar(redditUsernameEnvName)
	if envVarErr != nil {
		return shared.OutputSection{}, envVarErr
	}

	userAgent := fmt.Sprintf("golang Sponge:0.0.1 (by /u/%s)", redditUsername)
	golangListUrl := fmt.Sprintf("https://www.reddit.com/r/%s/top.json?raw_json=1&t=day&limit=%d", subName, numItems)

	// TODO: factor out client and json parsing
	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", golangListUrl, nil)
	req.Header.Set("User-Agent", userAgent) // required or reddit API will return 429 code
	resp, err := client.Do(req)
	if err != nil {
		return shared.OutputSection{}, err
	}
	if resp.StatusCode != 200 {
		return shared.OutputSection{}, fmt.Errorf("Non 200 response %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	redditList := redditList{}
	decoder := json.NewDecoder(resp.Body)
	if decodeErr := decoder.Decode(&redditList); decodeErr != nil {
		return shared.OutputSection{}, decodeErr
	}

	var redditItems []shared.StandardizedItem
	for _, item := range redditList.Data.Children {
		redditItems = append(redditItems, item.Data.standardize())
	}

	output := shared.OutputSection{
		Name:  fmt.Sprintf("Reddit r/%s", subName),
		Items: redditItems}

	return output, nil
}
