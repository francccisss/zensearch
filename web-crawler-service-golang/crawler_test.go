package main

import (
	"testing"
)

func TestStartCrawl(t *testing.T) {

	URLs := []string{"1", "2", "3", "4", "5", "6"}
	spawner := NewSpawner(10, URLs)
	crawlerCount := spawner.SpawnCrawlers()
	if crawlerCount != len(URLs) {
		t.Fatalf("Should return the same int as the length of the url array.")
	}

}
