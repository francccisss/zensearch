package selenium

import (
	"fmt"
	"log"
	"os"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

const (
	chromeDriverPath = "pkg/chrome/chromedriver"
	port             = 4444
	webDriverURL     = "http://localhost:%d/wd/hub"
)

const (
	linkEl = "a"
	para   = "p"
	span   = "span"
	code   = "code"
	pre    = "pre"
	h1     = "h1"
	h2     = "h2"
	h3     = "h3"
	h4     = "h4"
)

/*
Starts up new Chrome Driver server to handle requests via http from a remote
client (this program) to the Web browser's devtools using WebDriver protocol.
*/
func CreateWebDriverServer() error {
	opts := []selenium.ServiceOption{
		selenium.StartFrameBuffer(),
		selenium.ChromeDriver(chromeDriverPath),
		selenium.Output(os.Stderr),
	}
	service, err := selenium.NewChromeDriverService(chromeDriverPath, port, opts...)
	defer service.Stop()
	if err != nil {
		fmt.Printf(err.Error())
		return fmt.Errorf("Unable to create a Web driver server.")
	}
	log.Printf("Web Driver Server Created.\n")
	return fmt.Errorf("Unable to create a Web driver server.")
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
	wd, err := selenium.NewRemote(caps, fmt.Sprintf(webDriverURL, port))
	if err != nil {
		return nil, fmt.Errorf("ERROR: Unable to create a new remote client session with web driver server.")
	}
	log.Printf("Connected to Web Driver Server\n")
	return &wd, nil
}
