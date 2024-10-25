package main

import (
	"log"
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

	Docs := []string{"https://www.ringcentral.com/ccaas.html"}
	spawner := NewSpawner(10, Docs)
	spawner.SpawnCrawlers()

}
