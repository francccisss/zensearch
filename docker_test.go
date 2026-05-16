package main

import (
	"context"
	"testing"
)

const DOCKER_DESKTOP_HOST = "unix:///home/francois/.docker/desktop/docker.sock"

func TestWithExistingContainer(t *testing.T) {

	dm, err := NewDockerManager(DOCKER_DESKTOP_HOST)
	if err != nil {
		t.Fatalf("[TEST]: Existing Containers Test - %s\n", err)
	}
	ctx := context.Background()

	err = dm.PullImage(ctx, SeleniumConfig.ImageName, SeleniumConfig.Tag)

	if err != nil {
		t.Fatalf("[TEST]: Pulling Docker Image - %s\n", err)
	}
	dm.NewDockerContainer(SeleniumConfig)

	err = <-dm.Run(ctx, SeleniumConfig.Name, SeleniumConfig.ImageName, SeleniumConfig.Tag)

	if err != nil {
		t.Fatalf("[TEST]: Existing Containers Test - %s\n", err)
	}

	err = dm.RemoveContainer(ctx, SeleniumConfig.Name)
	if err != nil {
		t.Fatalf("[TEST]: Existing Containers Test - %s\n", err)
	}
}
