package main

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/docker/docker/client"
)

func TestNoContainer(t *testing.T) {

	ctx := context.Background()
	fmt.Printf("Docker: connecting client to docker daemon...\n")
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Panic(err.Error())
	}
	clientContainer := ClientContainer{
		Client:         cli,
		ContainerName:  "zensearch-cli-rabbitmq",
		HostPorts:      HostPorts{"5672", "15672"},
		ContainerPorts: ContainerPorts{{"5672", "5672"}, {"15672", "15672"}}}
	defer cli.Close()

	err = clientContainer.Start(ctx)
	if err != nil {
		fmt.Println(err.Error())
	}

	err = clientContainer.Stop(ctx)
	if err != nil {
		fmt.Println(err.Error())
	}
}
