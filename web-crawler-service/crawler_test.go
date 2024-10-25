package main

import (
	"fmt"
	"slices"
	"testing"
)

type pageNavigator struct {
	currentUrl   string
	pagesVisited map[string]string
	queue        Queue
}

func TestTraversal(t *testing.T) {

	pn := pageNavigator{
		currentUrl:   webpages[0].Url,
		pagesVisited: map[string]string{},
		queue: Queue{
			array: []string{webpages[0].Url},
		},
	}

	fmt.Println(pn.queue)
	pn.queue.enqueue("new item")
	fmt.Println(pn.queue)
	pn.queue.dequeue()
	fmt.Println(pn.queue)
	// err := pn.navigatePages()
	// if err != nil {
	// 	t.Fatalf("Why error")
	// }
}

type Queue struct {
	array []string
}

func (q *Queue) enqueue(item string) []string {
	q.array = append(q.array, item)
	return q.array
}

func (q *Queue) dequeue() string {
	head := q.array[0]
	slices.Delete(q.array, 0, 1)
	return head
}

func (pn *pageNavigator) navigatePages() error {

	// might be difficult if the current page has children that have not been traversed yet
	if _, visited := pn.pagesVisited[pn.currentUrl]; visited {
		return nil
	}

	return nil
}

type Webpage struct {
	Title string
	Url   string
	Links []string
}

var webpages = []Webpage{
	{
		Title: "Page 1",
		Url:   "https://example.com/page1",
		Links: []string{"https://example.com/link1", "https://example.com/link2"},
	},
	{
		Title: "Page 2",
		Url:   "https://example.com/page2",
		Links: []string{"https://example.com/link3", "https://example.com/link4"},
	},
	{
		Title: "Page 3",
		Url:   "https://example.com/page3",
		Links: []string{"https://example.com/link5", "https://example.com/link6"},
	},
	{
		Title: "Page 4",
		Url:   "https://example.com/page4",
		Links: []string{"https://example.com/link7", "https://example.com/link8"},
	},
	{
		Title: "Page 5",
		Url:   "https://example.com/page5",
		Links: []string{"https://example.com/link9", "https://example.com/link10"},
	},
	{
		Title: "Page 6",
		Url:   "https://example.com/page6",
		Links: []string{"https://example.com/link11", "https://example.com/link12"},
	},
	{
		Title: "Page 7",
		Url:   "https://example.com/page7",
		Links: []string{"https://example.com/link13", "https://example.com/link14"},
	},
	{
		Title: "Page 8",
		Url:   "https://example.com/page8",
		Links: []string{"https://example.com/link15", "https://example.com/link16"},
	},
	{
		Title: "Page 9",
		Url:   "https://example.com/page9",
		Links: []string{"https://example.com/link17", "https://example.com/link18"},
	},
	{
		Title: "Page 10",
		Url:   "https://example.com/page10",
		Links: []string{"https://example.com/link19", "https://example.com/link20"},
	},
}
