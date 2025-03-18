package main

import (
	"crawler/internal/rabbitmq"
	"crawler/internal/types"
	"fmt"
	"testing"
)

func TestCrawlerNotif(t *testing.T) {

	err := rabbitmq.EstablishConnection(7)

	if err != nil {
		fmt.Println(err.Error())
		t.Fatal(err)
	}
	conn, err := rabbitmq.GetConnection("conn")
	if err != nil {
		fmt.Println("Connection does not exist")
		t.Fatal(err)
	}
	fmt.Println("Crawler established TCP Connection with RabbitMQ")

	defer conn.Close()

	dbChannel, err := conn.Channel()
	if err != nil {
		fmt.Printf("Unable to create a crawl channel.\n")
	}
	rabbitmq.SetNewChannel("dbChannel", dbChannel)
	defer dbChannel.Close()

	result := types.IndexedResult{
		CrawlResult: types.CrawlResult{
			URLSeed:     "fzaid.vercel.app",
			Message:     "failed to crawl URLSeed",
			CrawlStatus: CRAWL_FAIL,
			TotalPages:  21,
		},
		Webpages: []types.IndexedWebpage{
			{
				Header: types.Header{
					Title: "menu",
					URL:   "fzaid.vercel.app/menu",
				},
				Contents: "Doobeedobeedapdap",
			},
		},
	}
	err = SendResults(result)
	if err != nil {
		t.Fatal(err)
	}
}
