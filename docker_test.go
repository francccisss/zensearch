package main

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestWithExistingContainer(t *testing.T) {

	newDockerClient, err := NewDockerClient()
	if err != nil {
		t.Fatalf("[TEST]: Existing Containers Test - %s\n", err)
	}
	ctx := context.Background()

	err = newDockerClient.PullImage(ctx, SeleniumConfig.ImageName, SeleniumConfig.Tag)

	if err != nil {
		t.Fatalf("[TEST]: Existing Containers Test - %s\n", err)
	}
	scont := NewDockerContainer(SeleniumConfig, &newDockerClient)

	err = <-scont.Run(ctx, "selenium/standalone-chrome", "latest")

	defer scont.Stop(ctx)

	if err != nil {
		t.Fatalf("[TEST]: Existing Containers Test - %s\n", err)
	}

	v := make(chan int)

	<-v
}
func TestNoContainer(t *testing.T) {

	dockerClient, err := NewDockerClient()
	ctx := context.Background()
	fmt.Printf("Docker: connecting client to docker daemon...\n")
	if err != nil {
		t.Fatalf("[TEST]: No Containers Test - %s\n", err)
	}
	dockerContainer := NewDockerContainer(RabbitmqConfig, &dockerClient)

	defer killCont(ctx, dockerContainer)

	err = dockerContainer.Start(ctx, "")

	if err != nil {
		t.Fatalf("[TEST]: No Containers Test - %s\n", err)
	}

	err = dockerContainer.Stop(ctx)
	if err != nil {
		t.Fatalf("[TEST]: No Containers Test - %s\n", err)
	}

}

func TestContainerStopAndStart(t *testing.T) {

	ctx := context.Background()

	dockerClient, err := NewDockerClient()
	if err != nil {
		t.Fatalf("[TEST]: Stop and Start - %s\n", err)
	}
	dockerContainer := NewDockerContainer(RabbitmqConfig, &dockerClient)
	defer killCont(ctx, dockerContainer)
	defer dockerClient.Client.Close()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		fmt.Println("Starting container")
		defer wg.Done()
		err := <-dockerContainer.Run(ctx, "rabbitmq", "4.0-management")
		if err != nil {
			t.Errorf("[TEST]: Stop and Start - %s\n", err)
		}
	}()

	fmt.Println("Docker: Sleeping for until container is running")
	wg.Wait()
	fmt.Println("Docker: Done sleeping, stopping container now...")

	err = dockerContainer.Stop(ctx)
	if err != nil {
		t.Fatalf("[TEST]: Stop and Start - %s\n", err)
	}

	fmt.Println("Docker: Sleeping for 2 seconds before starting")
	time.Sleep(time.Second * 4)
	fmt.Println("Docker: Done sleeping starting container")

	err = dockerContainer.Start(ctx, "")
	if err != nil {
		t.Fatalf("[TEST]: Stop and Start - %s\n", err)
	}
}

// When calling cancelFunc, and then docker.Stop, docker container does not stop
// for some reason, it could be cleaned up inside the docker sdk
func TestDockerStop(t *testing.T) {

	ctx, cancelFunc := context.WithCancel(context.Background())
	// ctx := context.Background()

	dockerClient, err := NewDockerClient()

	if err != nil {
		t.Fatalf("[TEST]: Docker Stop - %s\n", err)
	}
	cont := NewDockerContainer(RabbitmqConfig, &dockerClient)
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
	err := cc.Stop(ctx)
	if err != nil {
		fmt.Println("Docker: ERROR unable to stop container")
	}
	fmt.Println("Docker: test ending removing container")
}
