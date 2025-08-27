package main

import (
	"context"
	"fmt"
	"io"
	_ "io"
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
	Env     []string
}

type Client struct {
	Client *client.Client
	HostPorts
	ContainerPorts
	ContainerName string
	ContainerID   string
	ShmSize       int64
	Env           []string
}

type HostPorts []string
type ContainerPorts [][]string

// type Container interface {
// 	Run(dctx context.Context, imageName string, tag string) error
// 	Start(context.Context, string) error
// 	Stop(context.Context) error
// }

// TODO write context purpose of dtx

func NewContainer(name string, hports HostPorts, cports ContainerPorts, shmsize int64, contEnv []string) Client {
	fmt.Printf("Docker: connecting client to docker daemon...\n")
	var cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Panic(err.Error())
	}
	return Client{
		Client:         cli,
		ContainerName:  name,
		HostPorts:      hports,
		ContainerPorts: cports,
		ShmSize:        shmsize,
		Env:            contEnv,
	}
}

// TODO maybe instead of creating image name here, instead do it when creating a new Container?
// Pulls image from registry build the image, adds port mapping arguments and then runs
// the container (the port mapping is only done when the container first starts up)
func (cc *Client) Run(dctx context.Context, imageName string, tag string) <-chan error {

	// check architecture and OS

	fmt.Printf("%s: checking for existing container before running...\n", cc.ContainerName)
	errChan := make(chan error, 1)

	// This getter only searches for a <zensearch> specific container
	// eg: zensearch-rabbitmq, zensearch-mysql, zensearch-selenium-*

	c, exists := cc.getContainer(dctx)
	if exists {
		fmt.Printf("Docker: %s container already exist\n", cc.ContainerName)
		// is it running?
		err := cc.Start(dctx, c.ID)
		errChan <- err
		if err != nil {
			fmt.Printf("%s: unable to start from existing container...\n", cc.ContainerName)
			return errChan
		}
		// go cc.listenContainerState(dctx)
		fmt.Printf("%s: container exposed ports -> %+v\n", cc.ContainerName, cc.HostPorts)
		return errChan
	}

	fmt.Printf("%s: creating new container...\n", cc.ContainerName)
	// Creates a new zensearch specific container from an existing image
	imageNameWithTag := imageName + ":" + tag
	fmt.Printf("%s: pulling %s image...\n", cc.ContainerName, imageNameWithTag)
	reader, err := cc.Client.ImagePull(dctx, imageNameWithTag, image.PullOptions{})

	if err != nil {
		errChan <- err
		return errChan
	}

	_, err = io.Copy(os.Stdout, reader)
	if err != nil {
		fmt.Printf("%s: Reader ERROR: %s\n", cc.ContainerName, err.Error())
	}
	reader.Close()

	err = cc.create(dctx, imageName, tag)
	if err != nil {
		fmt.Printf("%s: Unable to create container\n", cc.ContainerName)
		errChan <- err
		return errChan
	}

	fmt.Printf("%s: starting container...\n", cc.ContainerName)

	if err := cc.Client.ContainerStart(dctx, cc.ContainerID, container.StartOptions{}); err != nil {
		fmt.Printf("%s: Unable to start container\n", cc.ContainerName)
		errChan <- err
		return errChan
	}

	// dont know when it is completely finished, need to set a timer for other
	// process that depends on rabbitmq

	// go cc.listenContainerState(dctx)
	fmt.Printf("%s: container started!\n", cc.ContainerName)
	fmt.Printf("%s: container exposed ports -> %+v\n", cc.ContainerName, cc.HostPorts)

	errChan <- nil
	return errChan
}

// TODO how to start already existing container?
// which means a container not created programmatically in here
// but instead passing in an existing ContainerID in the user's docker container list
func (cc *Client) Start(dctx context.Context, containerID string) error {

	fmt.Printf("%s: starting container...\n", cc.ContainerName)
	if containerID != "" {
		cc.ContainerID = containerID
		fmt.Printf("%s: assigning container ID for container\n", cc.ContainerName)
	}
	if cc.ContainerID == "" {
		return fmt.Errorf("%s: ERROR current container does not have an associated ContainerID which means the container does not exist, instead run the Run() function to create and run a new container from an image\n", cc.ContainerName)
	}

	err := cc.Client.ContainerStart(dctx, cc.ContainerID, container.StartOptions{})
	if err != nil {
		fmt.Printf("%s: Unable to start the container", cc.ContainerName)
		return err
	}
	fmt.Printf("%s: container started!\n", cc.ContainerName)

	return nil
}

// TODO POINTER NOT BEING UPDATED WHEN USING EXISTING CONTAINER
func (cc *Client) Stop(dctx context.Context) error {
	if cc.ContainerID == "" {
		return fmt.Errorf("%s: ERROR there's nothing to stop because the container does not exist\n", cc.ContainerName)
	}
	fmt.Printf("%s: stopping container...\n", cc.ContainerName)
	err := cc.Client.ContainerStop(dctx, cc.ContainerID, container.StopOptions{Signal: "SIGKILL"})
	if err != nil {
		fmt.Printf("%s: ERROR %s", cc.ContainerName, err)
		return fmt.Errorf("Docker: ERROR Something went wrong, zensearch is unable to stop the container %s with ID of %s\n", cc.ContainerID[:8], cc.ContainerName)
	}
	fmt.Printf("%s: Successfully stopped with ID starting with %s\n", cc.ContainerName, cc.ContainerID[:8])
	return nil
}

// Creates a new container and updates the cc's ContainerID field is successful
// else will panic dont use separately from Run() because port mapping is only initialized
// on container startup and not on creation... i dont know why
func (cc *Client) create(dctx context.Context, imageName string, tag string) error {
	fmt.Printf("%s: creating container...\n", cc.ContainerName)
	imageNameWithTag := imageName + ":" + tag
	fmt.Printf("%s: applying ports\n", cc.ContainerName)
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

	fmt.Printf("Docker: creating container from %s image as %s \n", imageNameWithTag, cc.ContainerName)

	// TODO ERROR from here for some reason
	resp, err := cc.Client.ContainerCreate(dctx, &container.Config{
		Image: imageNameWithTag,
		// attaching container to process exec is on `-it`
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		ExposedPorts: containerPorts,
		Env:          cc.Env,
	},
		&container.HostConfig{
			ShmSize: cc.ShmSize,
			Binds: []string{
				"/var/run/docker.sock:/var/run/docker.sock",
			},
			PortBindings: hostPorts}, nil, nil, cc.ContainerName)

	if err != nil {
		fmt.Print("Invalid reference format error from here\n")
		return err
	}
	fmt.Printf("Docker: %s's container ID %s\n", cc.ContainerName, resp.ID)
	cc.ContainerID = resp.ID
	return nil
}

// TODO figure out what to do with this
func (cc *Client) listenContainerState(dctx context.Context) {
	fmt.Printf("\n%s: waiting for container status...\n", cc.ContainerName)
	statusCh, errCh := cc.Client.ContainerWait(dctx, cc.ContainerID, container.WaitConditionNotRunning)
	// Listening to stdout of container
	go func() {
		out, err := cc.Client.ContainerLogs(dctx, cc.ContainerID, container.LogsOptions{ShowStdout: true, ShowStderr: true})

		if err != nil {
			fmt.Errorf(err.Error())
			<-dctx.Done()
			return
		}

		_, err = stdcopy.StdCopy(os.Stdout, os.Stderr, out)
		if err != nil {
			fmt.Println(err.Error())
		}
	}()

	select {
	case err := <-errCh:
		fmt.Printf("%s: closing container\n", cc.ContainerName)
		fmt.Printf("%s: cause for closing container: %s\n", cc.ContainerName, err.Error())
		return
	case s := <-statusCh:
		fmt.Printf("Container %s status:\n", cc.ContainerName)
		if s.Error != nil {
			fmt.Println(s.Error)
			return
		}
		fmt.Printf("Docker: %s container closed gracefully\n", cc.ContainerName)
	}
}

// should check if a rabbitmq container already exists
func (cc *Client) getContainer(dctx context.Context) (container.Summary, bool) {
	filter := filters.NewArgs()
	filter.Add("name", cc.ContainerName)
	fmt.Println("WHAT %s", cc.ContainerName)

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
