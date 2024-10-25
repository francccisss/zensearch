package main

import "slices"

type Queue struct {
	array []string
}

func (q *Queue) Enqueue(item ...string) []string {
	q.array = append(q.array, item...)
	return q.array
}

func (q *Queue) Dequeue() string {
	head := q.array[0]
	slices.Delete(q.array, 0, 1)
	return head
}
