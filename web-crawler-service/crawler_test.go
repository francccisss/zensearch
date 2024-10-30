package main

import (
	"fmt"
	"log"
	"os"
	"testing"
	"web-crawler-service/pkg/webdriver"
)

func TestTraversal(t *testing.T) {

	service, err := webdriver.CreateWebDriverServer()
	defer service.Stop()
	if err != nil {
		log.Print("INFO: Retry web driver server or the application.\n")
		log.Print(err.Error())
	}

	Docs := os.Args[3:]
	fmt.Printf("\nTest Argument : %+v\n", os.Args[3:][0])
	spawner := NewSpawner(10, Docs)
	spawner.SpawnCrawlers()

}
