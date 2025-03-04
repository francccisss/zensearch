package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func TestWithExistingContainer(t *testing.T) {
	ctx := context.Background()
	cont := NewContainer("zensearch-cli-rabbitmq", rabbitmqContConfig.HostPorts, rabbitmqContConfig.ContainerPorts)
	err := cont.Run(ctx, "rabbitmq", "4.0-management")
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}
}
func TestNoContainer(t *testing.T) {

	var cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	ctx := context.Background()
	fmt.Printf("Docker: connecting client to docker daemon...\n")
	if err != nil {
		log.Panic(err.Error())
	}
	clientContainer := ClientContainer{
		Client:         cli,
		ContainerName:  "zensearch-cli-rabbitmq",
		HostPorts:      HostPorts{"5672", "15672"},
		ContainerPorts: ContainerPorts{{"5672", "5672"}, {"15672", "15672"}}}
	defer killCont(ctx, clientContainer)

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

	var cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	fmt.Printf("Docker: connecting client to docker daemon...\n")
	if err != nil {
		log.Panic(err.Error())
	}
	clientContainer := ClientContainer{
		Client:         cli,
		ContainerName:  "zensearch-cli-rabbitmq",
		HostPorts:      HostPorts{"5672", "15672"},
		ContainerPorts: ContainerPorts{{"5672", "5672"}, {"15672", "15672"}}}
	defer killCont(ctx, clientContainer)
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

	fmt.Println("Docker: Sleeping for 2 seconds before starting")
	time.Sleep(time.Second * 4)
	fmt.Println("Docker: Done sleeping starting container")

	err = clientContainer.Start(ctx, "")
	if err != nil {
		fmt.Println(err.Error())
	}
}

func killCont(ctx context.Context, cc ClientContainer) {
	err := cc.Client.ContainerRemove(ctx, cc.ContainerID, container.RemoveOptions{Force: true})
	if err != nil {
		fmt.Println("Docker: ERROR unable to kill and remove container")
	}
	fmt.Println("Docker: test ending removing container")
}
