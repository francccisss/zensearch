package main

import (
	"crawler/internal/types"
	"crawler/utilities"
	"fmt"
	"log"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/tebeka/selenium"
)

type PageNavigator struct {
	WD              *selenium.WebDriver
	PagesVisited    map[string]string
	Queue           Queue
	DisallowedPaths []string
	RetryCount      int
	RequestTime
	IndexedWebpages []types.IndexedWebpage
	Hostname        string
}

type RequestTime struct {
	interval  int
	mselapsed int
}

const (
	MAX_RETRIES = 7
	// removes links to web objects that does not return an html page.
	LINK_FILTERS = `a:not([href$=".zip"]):not([href$=".svg"]):not([href$=".scss"]):not([href$=".css"]):not([href$=".pdf"]):not([href$=".exe"]):not([href$=".jpg"]):not([href$=".png"]):not([href$=".tar.gz"]):not([href$=".rar"]):not([href$=".7z"]):not([href$=".mp3"]):not([href$=".mp4"]):not([href$=".mkv"]):not([href$=".tar"]):not([href$=".xz"]):not([href$=".msi"])`
)

func (pn *PageNavigator) navigatePageWithRetries(retries int, currentUrl string) error {
	startTimer := time.Now()

	if retries > 0 {
		err := (*pn.WD).Get(currentUrl)
		if err != nil {
			pn.mselapsed = 0
			fmt.Println(err.Error())
			fmt.Println("Retrying connection...")
			time.Sleep(5 * time.Second)
			return pn.navigatePageWithRetries(retries-1, currentUrl)
		}
		timeout := time.Now()
		pn.mselapsed = int(timeout.Sub(startTimer) / 1000000)
		return nil
	}
	return fmt.Errorf("ERROR: Unable to retrieve webpage after several retries\n")
}

func (pn *PageNavigator) isPathAllowed(path string) bool {

	for _, dapath := range pn.DisallowedPaths {
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

func (pn *PageNavigator) ProcessSeed(currentUrl string) error {

	// VISITING STAGE
	if pn.RetryCount >= MAX_RETRIES {
		return fmt.Errorf("Exceeded maximum retry count for this website, the crawler might be blocked while crawling Url: %s\nreturning...", pn.Hostname)
	}

	// this will only be true if links are exhausted from queue, not from when starting the crawl
	if len(pn.Queue.array) == 0 {
		fmt.Printf("NOTIF: Queue is empty.\n")
		return nil
	}
	// Oh and while I was debugging, i forgot to call Dequeue and kept wondering
	// why the first element is not being removed... almost an hour i guess before
	// i figured it out.
	pn.Queue.Dequeue()

	fmt.Printf("NOTIF: `%s` has popped from queue.\n", currentUrl)
	_, visited := pn.PagesVisited[currentUrl]
	if visited {
		// its so that we can grab unique links and append to children of the current page
		fmt.Printf("NOTIF: Page already visited\n\n")
		return nil
	}
	pn.requestDelay(2)
	err := pn.navigatePageWithRetries(MAX_RETRIES, currentUrl)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	pn.PagesVisited[currentUrl] = currentUrl

	fmt.Println("NOTIF: Page set to visited.")

	// FETCHING STAGE (fetching anchor elements)
	args := []interface{}{LINK_FILTERS}
	linksInterface, err := (*pn.WD).ExecuteScript(`return function (linkFilter){
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
		href, _, _ := strings.Cut(link, "#")
		childHostname, path, err := utilities.GetHostname(href)

		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		isAllowed := pn.isPathAllowed(path)
		if !isAllowed {
			continue
		}
		// enqueue links that have not been visited yet and that are the same as the hostname
		_, visited := pn.PagesVisited[href]
		// I KEEP ADDING THE SAME ELEMENTS IN THE QUEUE I DONT UNDERSTAND!!!!
		if !visited && childHostname == pn.Hostname {
			pn.Queue.Enqueue(href)
		}
	}
	fmt.Printf("NOTIF: Link Count in current url: %d\n", len(pageLinks))
	fmt.Printf("NOTIF: Queue Length: %d\n", len(pn.Queue.array))

	// INDEXING PHASE

	indexedWebpage, err := pn.Index()
	if err != nil {
		// then skip this page
		fmt.Printf("ERROR: Something went wrong, unable to index current webpage.\n")
		return err
	}

	fmt.Printf("NOTIF: page %s indexed\n", currentUrl)

	// SAVING PHASE
	fmt.Println("NOTIF: storing indexed page")
	result := types.IndexedResult{
		CrawlResult: types.CrawlResult{
			URLSeed:     currentUrl,
			Message:     "Successfully indexed and stored webpages",
			CrawlStatus: CRAWL_SUCCESS,
		},
		Webpage: indexedWebpage,
	}

	err = SendResults(result)
	if err != nil {
		return fmt.Errorf("Unable to send indexed result to database service\nreturning...")
	}
	fmt.Println("NOTIF: stored indexed webpage")
	// pn.IndexedWebpages = append(pn.IndexedWebpages, indexedWebpage)

	/*
	 no child to traverse to then return to caller, the caller function will
	 then go to its next child in the children array.
	*/

	// to stop the crawler entirely after multiple retries from navigation
	for _, next := range pn.Queue.array {

		// the `next` is the one to be dequeued after calling navigatePages()
		err := pn.ProcessSeed(next)
		/*
			if error occured from traversing or any error has occured
			increment counter, the RetryCount is the maximum tries for an error occur again,
			if it is too mauch tnen might be better to just throw an error instead of continuing the crawl
		*/
		if err != nil {
			fmt.Println(err.Error())
			pn.RetryCount++
			continue
		}
	}
	return nil
}
func (pt PageNavigator) Index() (types.IndexedWebpage, error) {

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
				textContents, err := extractTextContent(pt.WD, selector)
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
	title, err := (*pt.WD).Title()
	if err != nil {
		log.Printf("ERROR: No title for this page")
	}

	url, err := (*pt.WD).CurrentURL()
	if err != nil {
		log.Printf("ERROR: No url for this page")
	}

	newIndexedPage := types.IndexedWebpage{
		Contents: pageContents,
		Header: types.Header{
			URL:   url,
			Title: title,
		},
	}
	return newIndexedPage, nil
}

func extractTextContent(WD *selenium.WebDriver, selector string) ([]string, error) {
	elementTextContents := make([]string, 0, 10)
	elements, err := (*WD).FindElements(selenium.ByCSSSelector, selector)
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
