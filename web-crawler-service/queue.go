package main

import (
	"slices"
)

type Queue struct {
	array []string
}

func (q *Queue) Enqueue(item string) {
	// need to check for duplicates, this is bad because
	// this would take O(n), maybe i should put this in a map instead?
	for _, queuedItem := range q.array {
		if item == queuedItem {
			// dont add the item
			return
		}
	}
	q.array = append(q.array, item)
	return
}

func (q *Queue) Dequeue() {
	q.array = slices.Delete(q.array, 0, 1)
}
