package main

import (
	"flag"
	"fmt"
	"github.com/bobo333/sponge/shared"
	"github.com/bobo333/sponge/sources"
	"os"
	"path/filepath"
	"sync"
)

/*
   TODO:
   - tests
   - filter Hacker News if no url
*/

func main() {
	defaultOutput := filepath.Join(os.TempDir(), "sponge_out")
	numItems := flag.Int("count", 10, "Number of items to fetch from each source")
	outputLocation := flag.String("out", defaultOutput, "Output file")
	emailDestination := flag.String("email", "", "The email address to send results to")
	textOutput := flag.Bool("text", false, "Output in text form (default is html)")

	flag.Parse()

	newsApiSources := []string{
		"the-new-york-times",
		"the-wall-street-journal",
		"the-washington-post",
		"the-economist",
		"techcrunch",
	}

	sectionsToGet := []func(int) (shared.OutputSection, error){
		sources.GetHackerNews,
	}

	subredditsToGet := []string{
		"golang",
		"python",
		"sysadmin",
		"programming",
		"liverpoolfc",
	}

	var wg sync.WaitGroup
	receiverChannel := make(chan shared.OutputSection)

	for _, newsApiSource := range newsApiSources {
		wg.Add(1)
		newsApiSource := newsApiSource

		go func() {
			defer wg.Done()

			output, err := sources.GetNewsApiItems(newsApiSource, *numItems)
			if err != nil {
				fmt.Printf("%s\n", err)
			} else {
				receiverChannel <- output
			}
		}()
	}

	for _, fxn := range sectionsToGet {
		wg.Add(1)
		fxn := fxn

		go func() {
			defer wg.Done()

			output, err := fxn(*numItems)
			if err != nil {
				fmt.Printf("%s\n", err)
			} else {
				receiverChannel <- output
			}
		}()
	}

	for _, subName := range subredditsToGet {
		wg.Add(1)
		subName := subName

		go func() {
			defer wg.Done()

			output, err := sources.GetReddit(subName, *numItems, sources.SubredditUrlMaker)
			if err != nil {
				fmt.Printf("%s\n", err)
			} else {
				receiverChannel <- output
			}
		}()
	}

	go func() {
		wg.Wait()
		close(receiverChannel)
	}()

	var sections []shared.OutputSection
	for section := range receiverChannel {
		sections = append(sections, section)
	}

	// turn sections into text or html output
	var rawOutput string
	if *textOutput {
		rawOutput = shared.CreateRawText(sections)
	} else {
		rawOutput = shared.CreateRawHtml(sections)
	}

	// write output to destination (file or email)
	if *emailDestination != "" {
		shared.WriteToEmail(*emailDestination, rawOutput, *textOutput)
	} else {
		shared.WriteToFile(*outputLocation, rawOutput, *textOutput)
	}
}
