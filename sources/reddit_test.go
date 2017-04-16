package sources_test

import "fmt"
import "io/ioutil"
import shared "github.com/bobo333/sponge/shared"
import sources "github.com/bobo333/sponge/sources"
import "net/http"
import "net/http/httptest"
import "os"
import "testing"

func TestGetReddit(t *testing.T) {
	bytes, _ := ioutil.ReadFile("reddit.txt")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, string(bytes))
	}))
	defer ts.Close()

	// mock the url maker so it uses the test server url
	urlMaker := func(_ string, _ int) string {
		return ts.URL
	}

	os.Setenv("REDDIT_USERNAME", "a_username")
	subName := "python"
	numItems := 2
	outputSection, err := sources.GetReddit(subName, numItems, urlMaker)
	if err != nil {
		t.Error("got error on reddit request")
	}

	expectedOutputName := fmt.Sprintf("Reddit r/%s", subName)
	if outputSection.Name != expectedOutputName {
		t.Errorf("expected output section to be %s, got %s", expectedOutputName, outputSection.Name)
	}
	if len(outputSection.Items) != numItems {
		t.Errorf("expected %d items, got %d", numItems, len(outputSection.Items))
	}

	expectedValues := []map[string]string{
		{
			"title":    "/r/Python Official Job Board",
			"url":      "https://www.reddit.com/r/Python/comments/6302cj/rpython_official_job_board/",
			"comments": "https://reddit.com/r/Python/comments/6302cj/rpython_official_job_board/"},
		{
			"title":    "Minimal and clean examples of data structures and algorithms in Python",
			"url":      "https://github.com/keon/algorithms",
			"comments": "https://reddit.com/r/Python/comments/65mxbw/minimal_and_clean_examples_of_data_structures_and/"}}

	verifyOutputItems(outputSection, expectedValues, t)
}

func TestSubredditUrlMaker(t *testing.T) {
	subName := "a_subreddit"
	numItems := 5

	expectedUrl := fmt.Sprintf("https://www.reddit.com/r/%s/top.json?raw_json=1&t=day&limit=%d", subName, numItems)
	actualUrl := sources.SubredditUrlMaker(subName, numItems)

	if actualUrl != expectedUrl {
		t.Errorf("expected url %s got %s", expectedUrl, actualUrl)
	}
}

func verifyOutputItems(outputSection shared.OutputSection,
	expectedValues []map[string]string, t *testing.T) {
	for i, item := range outputSection.Items {
		title, expectedTitle := item.Title, expectedValues[i]["title"]
		if title != expectedTitle {
			t.Errorf("expected title to be %s, got %s", expectedTitle, title)
		}
		url, expectedUrl := item.Url, expectedValues[i]["url"]
		if url != expectedUrl {
			t.Errorf("expected url to be %s, got %s", expectedUrl, url)
		}
		comments, expectedComments := item.Comments, expectedValues[i]["comments"]
		if comments != expectedComments {
			t.Errorf("expected comments to be %s, got %s", expectedComments, comments)
		}
	}
}
