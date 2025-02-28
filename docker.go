package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
)

type ClientContainer struct {
	Client *client.Client
	HostPorts
	ContainerPorts
	ContainerName string
	ContainerID   string
}

type HostPorts []string
type ContainerPorts [][]string

type Container interface {
	Run(ctx context.Context, imageName string, tag string)
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// Pulls image from registry build the image, adds port mapping arguments and then runs
// the container (the port mapping is only done when the container first starts up)
func (cc *ClientContainer) Run(ctx context.Context, imageName string, tag string) {

	fmt.Println("Docker: starting up container...")
	imageNameWithTag := imageName + ":" + tag
	fmt.Printf("Docker: pulling %s image...\n", imageNameWithTag)
	reader, err := cc.Client.ImagePull(ctx, imageName+":"+tag, image.PullOptions{})
	if err != nil {
		log.Panic(err)
	}

	io.Copy(os.Stdout, reader)
	defer reader.Close()

	cc.create(ctx, imageName, tag)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Docker: starting %s container...\n", cc.ContainerName)

	if err := cc.Client.ContainerStart(ctx, cc.ContainerID, container.StartOptions{}); err != nil {
		fmt.Println("Unable to start rabbitmq container")
		panic(err)
	}
	// dont know when it is completely finished, need to set a timer for other
	// process that depends on rabbitmq
	fmt.Printf("Docker: %s container started!\n", cc.ContainerName)
	fmt.Printf("Docker: %s container exposed ports -> %+v\n", cc.ContainerName, cc.HostPorts)

	fmt.Printf("Docker: waiting for %s container status...\n", cc.ContainerName)
	statusCh, errCh := cc.Client.ContainerWait(ctx, cc.ContainerID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		fmt.Println("Container state not positive")
		panic(err)
	case s := <-statusCh:
		fmt.Println("Container status:")
		if s.Error == nil {
			fmt.Println("Docker: container closed gracefully")
		}
	}
	out, _ := cc.Client.ContainerLogs(ctx, cc.ContainerID, container.LogsOptions{ShowStdout: true, ShowStderr: true})
	stdcopy.StdCopy(os.Stdout, os.Stderr, out)
}

func (cc *ClientContainer) Start(ctx context.Context) error {
	if cc.ContainerID == "" {
		return fmt.Errorf("Docker: ERROR current container does not have an associated ContainerID which means the container does not exist, instead run the Run() function to create and run a new container from an image")
	}
	err := cc.Client.ContainerStart(ctx, cc.ContainerID, container.StartOptions{})
	if err != nil {
		fmt.Println("Docker: Unable to start the container")
		return err
	}
	out, err := cc.Client.ContainerLogs(ctx, cc.ContainerID, container.LogsOptions{ShowStderr: true, ShowStdout: true})
	io.Copy(os.Stdout, out)

	return nil
}

func (cc *ClientContainer) Stop(ctx context.Context) error {
	if cc.ContainerID == "" {
		return fmt.Errorf("Docker: ERROR there's nothing to stop because the container does not exist")
	}
	err := cc.Client.ContainerStop(ctx, cc.ContainerID, container.StopOptions{})
	if err != nil {
		return fmt.Errorf("Docker: ERROR Something went wrong, zensearch is unable to stop the currently running container with ID starting with %s", cc.ContainerID[:8])
	}
	fmt.Printf("Docker: Successfully stopped %s with ID starting with %s", cc.ContainerName, cc.ContainerID[:8])
	return nil
}

// Creates a new container and updates the clientContainer's ContainerID field is successful
// else will panic dont use separately from Run() because port mapping is only initialized
// on container startup and not on creation... i dont know why
func (cc *ClientContainer) create(ctx context.Context, imageName string, tag string) {

	fmt.Println("Creating container")
	imageNameWithTag := imageName + ":" + tag
	hostPorts := map[nat.Port][]nat.PortBinding{}
	for _, hostPort := range cc.HostPorts {
		fmt.Println("Docker: applying host ports")
		_, ok := hostPorts[nat.Port(fmt.Sprintf("%s/tcp", hostPort))]
		if !ok {
			ports := hostPorts[nat.Port(fmt.Sprintf("%s/tcp", hostPort))]
			fmt.Printf("Current ports %+v\n", ports)
			hostPorts[nat.Port(fmt.Sprintf("%s/tcp", hostPort))] = append(ports, nat.PortBinding{HostIP: "0.0.0.0", HostPort: hostPort})
			ports = hostPorts[nat.Port(fmt.Sprintf("%s/tcp", hostPort))]
			fmt.Printf("Appended ports %+v\n", ports)
		}
	}
	containerPorts := map[nat.Port]struct{}{}
	for _, contPort := range cc.ContainerPorts {
		fmt.Println("Docker: applying container ports")
		_, ok := hostPorts[nat.Port(fmt.Sprintf("%s/tcp", contPort[0]))]
		if !ok {
			containerPorts[nat.Port(fmt.Sprintf("%s/tcp", contPort[0]))] = struct{}{}
		}
	}

	fmt.Println(cc.HostPorts)
	fmt.Println(cc.ContainerPorts)

	// grabs latest version of rabbitmq
	fmt.Printf("Docker: creating container from %s image as %s \n", imageNameWithTag, cc.ContainerName)
	resp, err := cc.Client.ContainerCreate(ctx, &container.Config{
		Image: imageNameWithTag,
		// attaching container to process exec is on `-it`
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		ExposedPorts: containerPorts,
	},
		&container.HostConfig{
			Binds: []string{
				"/var/run/docker.sock:/var/run/docker.sock",
			},
			PortBindings: hostPorts}, nil, nil, cc.ContainerName)

	if err != nil {
		fmt.Printf("Docker: ERROR was not able to create %s container", cc.ContainerName)
		panic(err)
	}
	fmt.Printf("Docker: %s's container ID %s\n", cc.ContainerName, resp.ID)
	cc.ContainerID = resp.ID
}

// Returnes specific container using filter to isolate container name
// used for checking duplicate containers
func (cc ClientContainer) getContainer(ctx context.Context) (container.Summary, error) {
	filter := filters.NewArgs()
	filter.Add("name", cc.ContainerName)

	containers, err := cc.Client.ContainerList(ctx, container.ListOptions{Size: false, Filters: filter, All: true})
	if err != nil {
		fmt.Println("Unable to get list of containers")
		panic(err)
	}

	if len(containers) == 0 {
		return container.Summary{}, fmt.Errorf("container %s does not exist", cc.ContainerName)
	}
	return containers[0], nil

}
