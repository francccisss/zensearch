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

type DockerContainerConfig struct {
	ImageName string
	Tag       string
	Name      string
	HostPorts
	ContainerPorts
	ShmSize int64
}

type ClientContainer struct {
	Client *client.Client
	HostPorts
	ContainerPorts
	ContainerName string
	ContainerID   string
	ShmSize       int64
}

type HostPorts []string
type ContainerPorts [][]string

// type Container interface {
// 	Run(dctx context.Context, imageName string, tag string) error
// 	Start(context.Context, string) error
// 	Stop(context.Context) error
// }

// TODO write context purpose of dtx

func NewContainer(name string, hports HostPorts, cports ContainerPorts, shmsize int64) ClientContainer {
	fmt.Printf("Docker: connecting client to docker daemon...\n")
	var cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Panic(err.Error())
	}
	return ClientContainer{
		Client:         cli,
		ContainerName:  name,
		HostPorts:      hports,
		ContainerPorts: cports,
		ShmSize:        shmsize,
	}
}

// TODO maybe instead of creating image name here, instead do it when creating a new Container?
// Pulls image from registry build the image, adds port mapping arguments and then runs
// the container (the port mapping is only done when the container first starts up)
func (cc *ClientContainer) Run(dctx context.Context, imageName string, tag string) error {

	fmt.Println("Docker: checking for existing container before running...")
	// err is nil if it exists, else not nil if container does not exists
	c, exists := cc.getContainer(dctx)
	if exists {
		// is it running?
		err := cc.Start(dctx, c.ID)
		if err != nil {
			fmt.Printf("Docker: unable to start %s from existing container...\n", cc.ContainerName)
			return err
		}
		go cc.listenContainerState(dctx)
		return nil
	}
	fmt.Println("Docker: creating container...")
	imageNameWithTag := imageName + ":" + tag
	fmt.Printf("Docker: pulling %s image...\n", imageNameWithTag)
	reader, err := cc.Client.ImagePull(dctx, imageName+":"+tag, image.PullOptions{})
	if err != nil {
		return err
	}

	io.Copy(os.Stdout, reader)
	defer reader.Close()
	err = cc.create(dctx, imageName, tag)
	if err != nil {
		fmt.Printf("Unable to create %s container", cc.ContainerName)
		return err
	}
	fmt.Printf("Docker: starting %s container...\n", cc.ContainerName)

	if err := cc.Client.ContainerStart(dctx, cc.ContainerID, container.StartOptions{}); err != nil {
		fmt.Printf("Unable to start %s container", cc.ContainerName)
		return err
	}
	// dont know when it is completely finished, need to set a timer for other
	// process that depends on rabbitmq

	fmt.Printf("Docker: %s container started!\n", cc.ContainerName)
	fmt.Printf("Docker: %s container exposed ports -> %+v\n", cc.ContainerName, cc.HostPorts)
	go cc.listenContainerState(dctx)
	return nil
}

// TODO how to start already existing container?
// which means a container not created programmatically in here
// but instead passing in an existing ContainerID in the user's docker container list
func (cc *ClientContainer) Start(dctx context.Context, containerID string) error {

	fmt.Printf("Docker: starting %s container...\n", cc.ContainerName)
	if containerID != "" {
		cc.ContainerID = containerID
		fmt.Printf("Docker: assigning container ID for %s\n", cc.ContainerName)
	}
	if cc.ContainerID == "" {
		return fmt.Errorf("Docker: ERROR current container does not have an associated ContainerID which means the container does not exist, instead run the Run() function to create and run a new container from an image\n")
	}

	err := cc.Client.ContainerStart(dctx, cc.ContainerID, container.StartOptions{})
	if err != nil {
		fmt.Println("Docker: Unable to start the container")
		return err
	}
	fmt.Printf("Docker: container %s started\n", cc.ContainerName)

	return nil
}

// TODO POINTER NOT BEING UPDATED WHEN USING EXISTING CONTAINER
func (cc *ClientContainer) Stop(dctx context.Context) error {
	if cc.ContainerID == "" {
		return fmt.Errorf("Docker: ERROR there's nothing to stop because the container does not exist\n")
	}
	fmt.Printf("Docker: stopping %s container...\n", cc.ContainerName)
	err := cc.Client.ContainerStop(dctx, cc.ContainerID, container.StopOptions{Signal: "SIGKILL"})
	if err != nil {
		return fmt.Errorf("Docker: ERROR Something went wrong, zensearch is unable to stop the currently running container with ID starting with %s\n", cc.ContainerID[:8])
	}
	fmt.Printf("Docker: Successfully stopped %s with ID starting with %s\n", cc.ContainerName, cc.ContainerID[:8])
	return nil
}

// Creates a new container and updates the cc's ContainerID field is successful
// else will panic dont use separately from Run() because port mapping is only initialized
// on container startup and not on creation... i dont know why
func (cc *ClientContainer) create(dctx context.Context, imageName string, tag string) error {

	fmt.Println("Docker: creating container...")
	imageNameWithTag := imageName + ":" + tag
	fmt.Println("Docker: applying ports")
	hostPorts := map[nat.Port][]nat.PortBinding{}
	for _, hostPort := range cc.HostPorts {
		_, ok := hostPorts[nat.Port(fmt.Sprintf("%s/tcp", hostPort))]
		if !ok {
			ports := hostPorts[nat.Port(fmt.Sprintf("%s/tcp", hostPort))]
			hostPorts[nat.Port(fmt.Sprintf("%s/tcp", hostPort))] = append(ports, nat.PortBinding{HostIP: "0.0.0.0", HostPort: hostPort})
			ports = hostPorts[nat.Port(fmt.Sprintf("%s/tcp", hostPort))]
		}
	}
	containerPorts := map[nat.Port]struct{}{}
	for _, contPort := range cc.ContainerPorts {
		_, ok := hostPorts[nat.Port(fmt.Sprintf("%s/tcp", contPort[0]))]
		if !ok {
			containerPorts[nat.Port(fmt.Sprintf("%s/tcp", contPort[0]))] = struct{}{}
		}
	}

	// grabs latest version of rabbitmq
	fmt.Printf("Docker: creating container from %s image as %s \n", imageNameWithTag, cc.ContainerName)
	resp, err := cc.Client.ContainerCreate(dctx, &container.Config{
		Image: imageNameWithTag,
		// attaching container to process exec is on `-it`
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		ExposedPorts: containerPorts,
	},
		&container.HostConfig{
			ShmSize: cc.ShmSize,
			Binds: []string{
				"/var/run/docker.sock:/var/run/docker.sock",
			},
			PortBindings: hostPorts}, nil, nil, cc.ContainerName)

	if err != nil {
		fmt.Printf("Docker: ERROR was not able to create %s container\n", cc.ContainerName)
		return err
	}
	fmt.Printf("Docker: %s's container ID %s\n", cc.ContainerName, resp.ID)
	cc.ContainerID = resp.ID
	return nil
}

// TODO figure out what to do with this
func (cc *ClientContainer) listenContainerState(dctx context.Context) {
	fmt.Printf("\nDocker: waiting for %s container status...\n", cc.ContainerName)
	statusCh, errCh := cc.Client.ContainerWait(dctx, cc.ContainerID, container.WaitConditionNotRunning)
	// Listening to stdout of container
	go func() {
		out, _ := cc.Client.ContainerLogs(dctx, cc.ContainerID, container.LogsOptions{ShowStdout: true, ShowStderr: true})
		stdcopy.StdCopy(os.Stdout, os.Stderr, out)
	}()
	select {
	case err := <-errCh:
		fmt.Println("docker: closing docker container")
		fmt.Printf("cause for closing container: %s\n", err.Error())
		return
	case s := <-statusCh:
		fmt.Println("Container status:")
		if s.Error != nil {
			fmt.Println(s.Error)
			return
		}
		fmt.Println("Docker: container closed gracefully")
	}

}

// Returnes specific container using filter to isolate container name
// used for checking duplicate containers
func (cc *ClientContainer) getContainer(dctx context.Context) (container.Summary, bool) {
	filter := filters.NewArgs()
	filter.Add("name", cc.ContainerName)

	containers, err := cc.Client.ContainerList(dctx, container.ListOptions{Size: false, Filters: filter, All: true})
	if err != nil {
		fmt.Println("Docker: ERROR Unable to get list of containers")
		return container.Summary{}, false
	}
	if len(containers) == 0 {
		fmt.Printf("Docker: container %s does not exist\n", cc.ContainerName)
		return container.Summary{}, false
	}
	fmt.Println(containers[0].ID)
	return containers[0], true

}
