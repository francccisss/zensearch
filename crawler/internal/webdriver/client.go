package webdriver

import (
	"fmt"
	"log"

	"github.com/tebeka/selenium"
)

const (
	webDriverURL = "http://localhost:4444/wd/hub"
)

/*
 Creating new clients or new sessions, to be placed into separate threads
 for handling multiple websites from the user's CrawlList
*/

func CreateClient() (*selenium.WebDriver, error) {
	caps := selenium.Capabilities{"browserName": "chromium", "goog:ChromeOptions": map[string]interface{}{
		"args": []string{
			"--headless",
			"--remote-debugging-pipe",
			"--no-sandbox",
			"disable-gpu",
		},
	}}
	wd, err := selenium.NewRemote(caps, webDriverURL)
	if err != nil {
		log.Print(err.Error())
		return nil, fmt.Errorf("ERROR: Unable to create a new remote client session with web driver server")
	}
	log.Printf("INFO: Client connected to Web Driver Server\n")
	return &wd, nil
}
