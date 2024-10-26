package main

import (
	"fmt"
	"log"
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
	RequestTime
}

type RequestTime struct {
	a         float32
	mselapsed int
}

func (pn *PageNavigator) navigatePageWithRetries(retries int, currentUrl string) error {
	startTimer := time.Now()

	if retries > 0 {
		err := (*pn.wd).Get(currentUrl)
		if err != nil {
			return pn.navigatePageWithRetries(retries-1, currentUrl)
		}

		timeout := time.Now()
		pn.mselapsed = int(timeout.Sub(startTimer) / 1000000)
		return nil
	}
	return fmt.Errorf("ERROR: Unable to retrieve webpage after several retries.")
}

func (pn *PageNavigator) isPathAllowed(path string) bool {
	for _, dapath := range pn.disallowedPaths {
		if strings.Contains(path, dapath) {
			fmt.Printf("Dapath: %s\n", dapath)
			return false
		}
	}
	return true
}

/*
TODO
Need to improve this, because whenever elapsed time is > min && elapsed < max
the value significantly reduces if it is faster compared to when elapsed > 1000ms
which is already slow and shouldnt be slowed down that much, im doing the opposite where
my goal was to make it sleep for longer if it is faster and dont sleep at all if it is
going too slow.
*/
func (pn *PageNavigator) requestDelay(multiplier int) {
	fmt.Printf("{elapsed: %d, threshold: %d}\n", pn.mselapsed, multiplier)
	min := 500
	max := 1500
	if pn.mselapsed < max {
		fmt.Printf("Too fast: %d\n", pn.mselapsed)
		time.Sleep(time.Duration((pn.mselapsed * 1000000) * multiplier))
	} else if pn.mselapsed < min {
		fmt.Printf("Too FAAAST: %d\n", pn.mselapsed)
		time.Sleep(time.Duration((pn.mselapsed * 1000000) * (multiplier * 3)))
	} else {
		fmt.Printf("Too slow: %d\n", pn.mselapsed)
		fmt.Printf("No Sleep\n")
	}
}

func (pn *PageNavigator) navigatePages(currentUrl string) error {

	fmt.Printf("NOTIF: Queue length %d.\n", len(pn.queue.array))

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
	pn.requestDelay(5)
	err := pn.navigatePageWithRetries(maxRetries, currentUrl)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	pn.pagesVisited[currentUrl] = currentUrl

	fmt.Println("NOTIF: Page set to visited.")
	fmt.Printf("NOTIF: Visited pages curent state %+v\n", pn.pagesVisited)

	pageLinks, err := (*pn.wd).FindElements(selenium.ByCSSSelector, linkFilter)

	// no children/error
	if err != nil {
		log.Println("ERROR: Unable to find elements of type `a`.")
		return fmt.Errorf("ERROR: Unable to find elements of type `a`.")
	}

	/*
		Need to check such that we can ignore the already visited links
		and use the ones that doesnt exist and consider it
		as the child of the currently visited link
	*/

	for _, link := range pageLinks {

		// need to filter out links that is not the same as hostname
		ref, _ := link.GetAttribute("href")
		cleanedRef, _, _ := strings.Cut(ref, "#")
		childHostname, path, err := utilities.GetHostname(cleanedRef)

		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		isAllowed := pn.isPathAllowed(path)
		if !isAllowed {
			fmt.Printf("Url path not allowed by robot txt: %s\n\n", path)
			continue
		}
		// enqueue links that have not been visited yet and that are the same as the hostname

		_, visited := pn.pagesVisited[cleanedRef]
		// I KEEP ADDING THE SAME ELEMENTS IN THE QUEUE I DONT UNDERSTAND!!!!
		if !visited && pn.entry.hostname == childHostname {
			fmt.Printf("Not visited yet: %s\n", cleanedRef)
			pn.queue.Enqueue(cleanedRef)
		}
	}

	indexedWebpage, err := pn.Index()
	if err != nil {
		// then skip this page
		fmt.Printf("ERROR: Something went wrong, unable to index current webpage.\n")
		return nil
	}
	pn.entry.IndexedWebpages = append(pn.entry.IndexedWebpages, indexedWebpage)

	/*
	 no child to traverse to then return to caller, the caller function will
	 then go to its next child in the children array.
	*/

	fmt.Printf("NOTIF: Current Queue State: %+v\n\n", pn.queue.array)
	for _, next := range pn.queue.array {

		err := pn.navigatePages(next)
		// if error occured from traversing or any error has occured
		// just move to the next child
		if err != nil {
			fmt.Printf("ERROR: Unable to navigate to `%s` just continue.\n", next)
			// im saying continue when length of queue is 0
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
	fmt.Printf("NOTIF: Page %s Indexed\n", url)

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
