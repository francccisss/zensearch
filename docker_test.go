package main

import (
	"context"
	"fmt"
	"testing"
)

const DOCKER_DESKTOP_HOST = "unix:///home/francois/.docker/desktop/docker.sock"

func TestWithExistingContainer(t *testing.T) {

	fmt.Println("[Test]: Running with existing container")

	dm, err := NewDockerManager(DOCKER_DESKTOP_HOST)
	if err != nil {
		t.Fatalf("[TEST]: Existing Containers Test - %s\n", err)
	}
	ctx := context.Background()

	fmt.Println("[Test]: Pulling image")
	err = dm.PullImage(ctx, SeleniumConfig.ImageName, SeleniumConfig.Tag)

	if err != nil {
		t.Fatalf("[TEST]: Pulling Docker Image - %s\n", err)
	}

	fmt.Println("[Test]: Creating new docker container")
	dm.NewDockerContainer(SeleniumConfig)

	_, exists := dm.GetContainerID(ctx, SeleniumConfig.Name)

	if exists {
		dm.RemoveContainer(ctx, SeleniumConfig.Name)
	}

	err = <-dm.Run(ctx, SeleniumConfig.Name, SeleniumConfig.ImageName, SeleniumConfig.Tag)

	if err != nil {
		t.Fatalf("[TEST]: Existing Containers Test - %s\n", err)
	}

	err = dm.RemoveContainer(ctx, SeleniumConfig.Name)
	if err != nil {
		t.Fatalf("[TEST]: Existing Containers Test - %s\n", err)
	}
}
