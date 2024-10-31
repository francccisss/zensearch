package main

import (
	"fmt"
	"log"
	"math"
	"strings"
	"sync"
	"time"
	"web-crawler-service/utilities"

	"github.com/tebeka/selenium"
)

type PageNavigator struct {
	entry           *WebpageEntry
	wd              *selenium.WebDriver
	pagesVisited    map[string]string
	queue           Queue
	disallowedPaths []string
	retryCount      int
	RequestTime
}

type RequestTime struct {
	interval  int
	mselapsed int
}

const (
	maxRetries = 7
	// removes links to web objects that does not return an html page.
	linkFilter = `a:not([href$=".zip"]):not([href$=".svg"]):not([href$=".scss"]):not([href$=".css"]):not([href$=".pdf"]):not([href$=".exe"]):not([href$=".jpg"]):not([href$=".png"]):not([href$=".tar.gz"]):not([href$=".rar"]):not([href$=".7z"]):not([href$=".mp3"]):not([href$=".mp4"]):not([href$=".mkv"]):not([href$=".tar"]):not([href$=".xz"]):not([href$=".msi"])`
)

func (pn *PageNavigator) navigatePageWithRetries(retries int, currentUrl string) error {
	startTimer := time.Now()

	if retries > 0 {
		err := (*pn.wd).Get(currentUrl)
		if err != nil {
			pn.mselapsed = 0
			fmt.Println(err.Error())
			return pn.navigatePageWithRetries(retries-1, currentUrl)
		}
		timeout := time.Now()
		pn.mselapsed = int(timeout.Sub(startTimer) / 1000000)
		return nil
	}
	return fmt.Errorf("ERROR: Unable to retrieve webpage after several retries\n")
}

func (pn *PageNavigator) isPathAllowed(path string) bool {

	// bro I only understand english :D just remove the ones that you want to be included
	languagePaths := []string{"es", "ko", "tr", "th", "it", "uk", "sk", "fr", "de", "zh", "ja", "ru", "ar", "pt", "hi"}
	pn.disallowedPaths = append(pn.disallowedPaths, languagePaths...)
	for _, dapath := range pn.disallowedPaths {
		if strings.Contains(path, dapath) {
			return false
		}
	}
	return true
}

/*
using elapsed time from start to end of request in milliseconds and compressing
it using log to smooth the values for increasing intervals for each requests
such that it doesnt grow too much when multiplying intervals.

multiplier values:
  - 0 ignores all intervals
  - 1 increases slowly but is still fast and might be blocked
  - 2 sweet middleground

The first check for pn.interval < min is hack i dont know what else to do.
*/
func (pn *PageNavigator) requestDelay(multiplier int) {
	max := 10000
	base := int(math.Log10(float64(pn.mselapsed)))

	fmt.Printf("CURRENT ELAPSED TIME: %d\n", pn.mselapsed)
	if pn.interval < max {
		pn.interval = (pn.interval + base) * multiplier
		fmt.Printf("INCREASE INTERVAL: %d\n", pn.interval)
	} else if pn.interval > max {
		fmt.Printf("RESET INTERVAL: %d\n", pn.interval)
		pn.interval = 0
	}
	time.Sleep(time.Duration(pn.interval * 1000000))
}

func (pn *PageNavigator) navigatePages(currentUrl string) error {

	if pn.retryCount >= maxRetries {
		return fmt.Errorf("Exceeded maximum retry count for this website, the crawler might be blocked while crawling Url: %s\nreturning...", pn.entry.hostname)
	}

	if len(pn.queue.array) == 0 {
		fmt.Printf("NOTIF: Queue is empty.\n")
		return nil
	}
	// Oh and while I was debugging, i forgot to call Dequeue and kept wondering
	// why the first element is not being removed... almost an hour i guess before
	// i figured it out.
	pn.queue.Dequeue()

	fmt.Printf("NOTIF: `%s` has popped from queue.\n", currentUrl)
	_, visited := pn.pagesVisited[currentUrl]
	if visited {
		// its so that we can grab unique links and append to children of the current page
		fmt.Printf("NOTIF: Page already visited\n\n")
		return nil
	}
	pn.requestDelay(2)
	err := pn.navigatePageWithRetries(maxRetries, currentUrl)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	pn.pagesVisited[currentUrl] = currentUrl

	fmt.Println("NOTIF: Page set to visited.")

	args := []interface{}{linkFilter}
	linksInterface, err := (*pn.wd).ExecuteScript(`return function (linkFilter){
    console.log(linkFilter)
    const links = document.querySelectorAll(linkFilter)
    return Array.from(links).map(link=>link.href)
    }(arguments[0])`, args)

	// no children/error
	if err != nil {
		log.Println("ERROR: Unable to find elements of type `a` something went wrong with the webdriver")
		return err
	}

	/*
	   Type assertions from the script that returns an interface{} which is an array
	   of filtered achor elements
	*/

	links, ok := linksInterface.([]interface{})
	if !ok {
		log.Println("ERROR: Failed to convert linksInterface to []interface{}")
		return fmt.Errorf("type assertion to []interface{} failed\n")
	}
	pageLinks := make([]string, len(links))
	for i, link := range links {
		if strLink, ok := link.(string); ok {
			pageLinks[i] = strLink
		} else {
			log.Printf("ERROR: Link at index %d is not a string, skipping this index\n", i)
		}
	}

	for _, link := range pageLinks {

		// need to filter out links that is not the same as hostname
		before, _, _ := strings.Cut(link, "#")
		childHostname, path, err := utilities.GetHostname(before)

		// For some odd reason, when using a selenium image the url paths of links does not resolve
		// as absolute path but instead returns relative path `/path/to/resource`

		// keep in mind if childHostname is "" then it is for sure a relative path

		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		isAllowed := pn.isPathAllowed(path)
		if !isAllowed {
			continue
		}
		// enqueue links that have not been visited yet and that are the same as the hostname
		_, visited := pn.pagesVisited[link]
		// I KEEP ADDING THE SAME ELEMENTS IN THE QUEUE I DONT UNDERSTAND!!!!
		if !visited && childHostname == pn.entry.hostname {
			pn.queue.Enqueue(link)
		}
	}
	fmt.Printf("NOTIF: Link Count in current url: %d\n", len(pageLinks))
	fmt.Printf("NOTIF: Queue Length: %d\n", len(pn.queue.array))
	indexedWebpage, err := pn.Index()
	if err != nil {
		// then skip this page
		fmt.Printf("ERROR: Something went wrong, unable to index current webpage.\n")
		return err
	}

	fmt.Printf("NOTIF: Page %s Indexed\n", currentUrl)
	pn.entry.IndexedWebpages = append(pn.entry.IndexedWebpages, indexedWebpage)

	/*
	 no child to traverse to then return to caller, the caller function will
	 then go to its next child in the children array.
	*/

	// to stop the crawler entirely after multiple retries from navigation
	for _, next := range pn.queue.array {

		// the `next` is the one to be dequeued after calling navigatePages()
		err := pn.navigatePages(next)
		/*
			if error occured from traversing or any error has occured
			increment counter, the retryCount is the maximum tries for an error occur again,
			if it is too mauch tnen might be better to just throw an error instead of continuing the crawl
		*/
		if err != nil {
			fmt.Println(err.Error())
			pn.retryCount++
			continue
		}
	}
	return nil
}
func (pt PageNavigator) Index() (IndexedWebpage, error) {

	/*
		Iterating through the elementSelector, where each selector, creates
		a new go routine, so using a buffered channel with the exact length of
		the indexSelector would make more sense.

		If ever we want to throttle the operation we can create a semaphore by
		limiting the buffered channel, if resource is an issue.
	*/

	htmlTextElementChan := make(chan string, len(elementSelector))
	var wg sync.WaitGroup

	// Start wait group after go routine is processed on a different thread
	wg.Add(1)

	// Go routine generator
	go func() {
		defer wg.Done()
		for _, selector := range elementSelector {
			wg.Add(1)
			go func(selector string) {
				defer wg.Done()
				textContents, err := extractTextContent(pt.wd, selector)
				if err != nil {
					htmlTextElementChan <- ""
					log.Print("ERROR: unable to extract text contents")
					return
				}

				// Joins the array of text contents and returns as a whole string of text content
				// from the current element.
				htmlTextElementChan <- joinTextContents(textContents)
			}(selector)
		}
	}()

	fmt.Println("NOTIF: Waiting for page indexer")
	wg.Wait()
	close(htmlTextElementChan)
	textChanSlice := make([]string, 0, 100)

	// for every joined text contents of each element on the current page,
	// append each block of text into a new array then join to represent
	// the whole contents of the page.
	for elementContents := range htmlTextElementChan {
		textChanSlice = append(textChanSlice, elementContents)
	}

	pageContents := joinTextContents(textChanSlice)
	title, err := (*pt.wd).Title()
	if err != nil {
		log.Printf("ERROR: No title for this page")
	}

	url, err := (*pt.wd).CurrentURL()
	if err != nil {
		log.Printf("ERROR: No url for this page")
	}

	newIndexedPage := IndexedWebpage{
		Contents: pageContents,
		Header: Header{
			Url:   url,
			Title: title,
		},
	}
	return newIndexedPage, nil
}

func extractTextContent(wd *selenium.WebDriver, selector string) ([]string, error) {
	elementTextContents := make([]string, 0, 10)
	elements, err := (*wd).FindElements(selenium.ByCSSSelector, selector)
	if err != nil {
		log.Printf("ERROR: Elements does not satisfy css selector: %s", selector)
		return nil, err
	}
	for _, el := range elements {
		text, err := el.Text()
		if err != nil {
			continue
		}
		elementTextContents = append(elementTextContents, text)
	}
	return elementTextContents, nil
}

func joinTextContents(tc []string) string {
	return strings.Join(tc, " ")
}
