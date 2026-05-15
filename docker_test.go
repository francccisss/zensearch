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

	newDockerClient, err := NewDockerClient()
	if err != nil {
		t.Fatalf("[TEST]: Existing Containers Test - %s\n", err)
	}
	ctx := context.Background()

	scont := NewDockerContainer(SeleniumContConfig, &newDockerClient)

	err = <-scont.Run(ctx, "selenium/standalone-chrome", "latest")

	defer scont.Stop(ctx)

	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}

	v := make(chan int)

	<-v
}
func TestNoContainer(t *testing.T) {

	dockerClient, err := NewDockerClient()
	ctx := context.Background()
	fmt.Printf("Docker: connecting client to docker daemon...\n")
	if err != nil {
		log.Panic(err.Error())
	}
	dockerContainer := NewDockerContainer(RabbitmqContConfig, &dockerClient)

	defer killCont(ctx, dockerContainer)

	err = dockerContainer.Start(ctx, "")
	if err != nil {
		fmt.Println(err.Error())
	}

	err = dockerContainer.Stop(ctx)
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
	dockerContainer := ClientContainer{
		Client:         cli,
		ContainerName:  "zensearch-cli-rabbitmq",
		HostPorts:      HostPorts{"5672", "15672"},
		ContainerPorts: ContainerPorts{{"5672", "5672"}, {"15672", "15672"}}}
	defer killCont(ctx, dockerContainer)
	defer cli.Close()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		fmt.Println("Starting container")
		defer wg.Done()
		err := dockerContainer.Run(ctx, "rabbitmq", "4.0-management")
		if err != nil {
			panic(err)
		}
	}()

	fmt.Println("Docker: Sleeping for until container is running")
	wg.Wait()
	fmt.Println("Docker: Done sleeping, stopping container now...")

	err = dockerContainer.Stop(ctx)
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println("Docker: Sleeping for 2 seconds before starting")
	time.Sleep(time.Second * 4)
	fmt.Println("Docker: Done sleeping starting container")

	err = dockerContainer.Start(ctx, "")
	if err != nil {
		fmt.Println(err.Error())
	}
}

// When calling cancelFunc, and then docker.Stop, docker container does not stop
// for some reason, it could be cleaned up inside the docker sdk
func TestDockerStop(t *testing.T) {

	ctx, cancelFunc := context.WithCancel(context.Background())
	// ctx := context.Background()

	cont := NewDockerContainer(RabbitmqContConfig)
	cont.Run(context.Background(), "rabbitmq", "4.0-management")

	sleepTime := 2
	fmt.Printf("TEST: Exiting in %d seconds\n", sleepTime)
	time.Sleep(time.Duration(sleepTime) * time.Second)
	fmt.Println("TEST: Exiting container")
	cancelFunc()
	//
	<-ctx.Done()
	fmt.Println("TEST: Stopped")
	cont.Stop(context.Background())
	v := make(chan int)
	<-v

}

func killCont(ctx context.Context, cc Container) {
	cc.Stop()
	err := cc.Client.ContainerRemove(ctx, cc.ContainerID, container.RemoveOptions{Force: true})
	if err != nil {
		fmt.Println("Docker: ERROR unable to kill and remove container")
	}
	fmt.Println("Docker: test ending removing container")
}
