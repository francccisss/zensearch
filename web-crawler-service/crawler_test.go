package main

import (
	"fmt"
	"os"
	"testing"
)

func TestTraversal(t *testing.T) {

	Docs := os.Args[3:]
	fmt.Printf("\nTest Argument : %+v\n", os.Args[3:][0])
	spawner := NewSpawner(10, Docs)
	spawner.SpawnCrawlers()

}
