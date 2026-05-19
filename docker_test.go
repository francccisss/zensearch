package main

import (
	"context"
	"fmt"
	"testing"
)

func TestWithExistingContainer(t *testing.T) {
	fmt.Println("[Test]: Running with existing container")
	fmt.Println("[TEST]: COMMAND - RUN ")

	dm, err := NewDockerManager()
	if err != nil {
		t.Fatalf("[TEST]: Existing Containers Test - %s\n", err)
	}
	ctx := context.Background()

	dm.NewDockerContainer(SeleniumConfig)

	err = dm.PullImage(ctx, SeleniumConfig.ImageName, SeleniumConfig.Tag)
	if err != nil {
		t.Fatalf("[TEST]: Image Pull Test - %s\n", err)
	}
	err = <-dm.Run(ctx, SeleniumConfig.Name, SeleniumConfig.ImageName, SeleniumConfig.Tag)

	if err != nil {
		t.Fatalf("[TEST]: Existing Containers Test - %s\n", err)
	}
	err = dm.Stop(ctx, SeleniumConfig.Name)
	if err != nil {
		t.Fatalf("[TEST]: Existing Containers Test - %s\n", err)
	}
}
