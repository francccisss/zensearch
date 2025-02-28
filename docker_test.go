package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

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

	err = clientContainer.Start(ctx, "")
	if err != nil {
		fmt.Println(err.Error())
	}

	err = clientContainer.Stop(ctx)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func TestContainerStopAndStart(t *testing.T) {

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

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		fmt.Println("Starting container")
		defer wg.Done()
		err := clientContainer.Run(ctx, "rabbitmq", "4.0-management")
		if err != nil {
			panic(err)
		}
	}()

	fmt.Println("Docker: Sleeping for until container is running")
	wg.Wait()
	fmt.Println("Docker: Done sleeping, stopping container now...")

	err = clientContainer.Stop(ctx)
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println("Docker: Sleeping for 10 seconds before starting")
	time.Sleep(time.Second * 10)
	fmt.Println("Docker: Done sleeping starting container")

	err = clientContainer.Start(ctx, "")
	if err != nil {
		fmt.Println(err.Error())
	}

}
