package webdriver

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

const (
	port         = 4444
	webDriverURL = "http://localhost:%d/wd/hub"
)

/*
Starts up new Chrome Driver server to handle requests via http from a remote
client (this program) to the Web browser's devtools using WebDriver protocol.
*/
func CreateWebDriverServer() (*selenium.Service, error) {
	path, _ := os.Getwd()
	chromeDriverPath := filepath.Join(path, "pkg", "webdriver", "chromedriver")
	fmt.Printf("Chrome Driver file path found: %s", chromeDriverPath)
	opts := []selenium.ServiceOption{
		selenium.StartFrameBuffer(),             // THIS USES THE DEPENDENCY X VIRTUAL FRAME BUFFER
		selenium.ChromeDriver(chromeDriverPath), // THIS USES THE DEPENDENCY CHROME WEB DRIVER
		// selenium.Output(os.Stderr),
	}
	service, err := selenium.NewChromeDriverService(chromeDriverPath, port, opts...)
	if err != nil {
		log.Print(err.Error())
		return nil, fmt.Errorf("ERROR: Unable to create a Web driver server")
	}
	log.Printf("INFO: Web Driver Server Created.\n")
	return service, nil
}

/*
 Creating new clients or new sessions, to be placed into separate threads
 for handling multiple websites from the user's CrawlList
*/

func CreateClient() (*selenium.WebDriver, error) {
	caps := selenium.Capabilities{"browserName": "chrome"}
	caps.AddChrome(chrome.Capabilities{Args: []string{
		"--headless",
		"--remote-debugging-pipe",
		"--no-sandbox",
		"disable-gpu",
	}})
	wd, err := selenium.NewRemote(caps, "")
	if err != nil {
		log.Print(err.Error())
		return nil, fmt.Errorf("ERROR: Unable to create a new remote client session with web driver server")
	}
	log.Printf("INFO: Client connected to Web Driver Server\n")
	return &wd, nil
}
