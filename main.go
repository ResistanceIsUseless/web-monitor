package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/projectdiscovery/goflags"
	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/gologger/levels"
	"github.com/projectdiscovery/httpx/runner"
)

// Options contains configuration options for command
// line flags and httpx settings
type Options struct {
	// general configuration options
	Verbose         bool
	FollowRedirects bool
	// input options
	addURL    string
	CheckURLs bool
}

// create json file to store results
type URLS struct {
	UniqueID                      string `json:"unique_id"`
	URL                           string `json:"url"`
	LastChecked                   string `json:"last_checked"`
	StatusCode                    int    `json:"status_code"`
	ContentLength                 int    `json:"content_length"`
	ContentType                   string `json:"content_type"`
	RedirectURL                   string `json:"redirect_url"`
	RedirectResponseCode          int    `json:"redirect_response_code"`
	RedirectResponseContentLength int    `json:"redirect_response_content_length"`
}

func main() {
	gologger.DefaultLogger.SetMaxLevel(levels.LevelVerbose)
	options := &Options{}
	flagParser := goflags.NewFlagParser(options)
	flagParser.AdditionalHelpContents = `
Example command: monitor -addURL https://www.google.com -CheckURLs true`
	flagParser.Parse()

	if options.CheckURLs {
		checkURLs()
		if options.addURL == "" {
			log.Fatal("Please provide a URL to check")
		}
	}

}
func httprunner(URL string) []URLS {
	// create a new httpx runner
	options := runner.Options{
		// general configuration options
		Verbose:         true,
		FollowRedirects: true,
		// input options
		addURL:    "https://www.google.com",
		CheckURLs: true,
	}
	runner, err := runner.New(options)
	if err != nil {
		log.Fatal(err)
	}

	// run the runner
	err = runner.RunEnumeration()
	if err != nil {
		log.Fatal(err)
	}

	// get the results
	results := runner.Results()
	for _, result := range results {
		return results
	}
}

// save results to json file
func saveResults(results []URLS) {

	var URLS = URLS{
		UniqueID:      "",
		URL:           "",
		LastChecked:   "",
		StatusCode:    0,
		ContentLength: 0,
		ContentType:   "",
	}

	// Marshal the modified struct back to JSON
	updatedJSON, err := json.MarshalIndent(URLS, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}

	// Write the JSON back to the file
	err = os.WriteFile("urls.json", updatedJSON, 0644)
	if err != nil {
		fmt.Println("Error writing JSON to file:", err)
		return
	}

}

func checkURLs() {
	file, err := os.ReadFile("urls.json")
	if err != nil {
		fmt.Println("Error reading JSON file:", err)
		return
	}

	// Unmarshal the JSON data into the struct
	var URLS []URLS
	err = json.Unmarshal(file, &URLS)
	if err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		return
	}
	// run httprunner for each URL in json file and compare and save results if changed
	for _, URL := range URLS {
		httprunner(URL.URL)
		results := httprunner(URL.URL)
		for _, result := range results {
			if result.ContentLength != URL.ContentLength {
				fmt.Println("Content length has changed")
				saveResults(results)
			} else {
				fmt.Println("Content length has not changed")
			}

		}
	}
}

// check if url is already in json file and if not add it
func addURL(URL string) {
	file, err := os.ReadFile("urls.json")
	if err != nil {
		fmt.Println("Error reading JSON file:", err)
		return
	}

	// Unmarshal the JSON data into the struct
	var URLS URLS
	err = json.Unmarshal(file, &URLS)
	if err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		return
	} else {
		fmt.Println("JSON file read successfully.")
	}

	// Check if the value exists
	B64ID := base64.StdEncoding.EncodeToString([]byte(URL))
	if URLS.UniqueID != B64ID {
		// The value doesn't exist, so add it
		URLS.URL = URL
		URLS.UniqueID = B64ID
		URLS.LastChecked = ""
		URLS.StatusCode = 0
		URLS.ContentLength = 0
		URLS.ContentType = ""

		// Marshal the modified struct back to JSON
		updatedJSON, err := json.MarshalIndent(URLS, "", "  ")
		if err != nil {
			fmt.Println("Error marshaling JSON:", err)
			return
		}

		// Write the JSON back to the file
		err = os.WriteFile("urls.json", updatedJSON, 0644)
		if err != nil {
			fmt.Println("Error writing JSON to file:", err)
			return
		}
		fmt.Println("New value added and file updated.")
	} else {
		fmt.Println("Value exists, no update needed.")
	}
}
