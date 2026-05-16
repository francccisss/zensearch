package main

import (
	"context"
	"fmt"
	"io"
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

type HostPorts []string
type ContainerPorts [][]string

type DockerClient struct {
	Client *client.Client
}

// TODO write context purpose of dtx

type DockerContainer struct {
	*DockerClient
	HostPorts
	ContainerPorts
	ContainerName string
	ContainerID   string
	ShmSize       int64
	Env           []string
	Tag           string
}

func NewDockerClient(dockerHost string) (DockerClient, error) {
	cl, err := client.NewClientWithOpts(client.FromEnv, client.WithHost(dockerHost))
	if err != nil {
		return DockerClient{}, nil
	}
	return DockerClient{cl}, nil

}

func NewDockerContainer(contConfig DockerContainerConfig, dockerClient *DockerClient) Container {
	fmt.Printf("[New Docker Container NOTICE]: Creating new docker container for %s\n", contConfig.Name)
	return &DockerContainer{
		DockerClient:   dockerClient,
		ContainerName:  contConfig.Name,
		HostPorts:      contConfig.HostPorts,
		ContainerPorts: contConfig.ContainerPorts,
		ShmSize:        contConfig.ShmSize,
		Env:            contConfig.Env,
		Tag:            contConfig.Tag,
	}
}

type Container interface {
	Run(dctx context.Context, imageName string, imageTag string) <-chan error
	Create(dctx context.Context, imageName string, tag string) error
	Start(dctx context.Context, containerID string) error
	Stop(dctx context.Context) error
	GetContainer(dctx context.Context) (container.Summary, bool)
	listenContainerState(dctx context.Context)
	GetName() string
	GetID() string
	GetTag() string
}

// Run() pulls an image from the docker registry given the container configuration
// created with NewDockerContainer,

func (dClient *DockerClient) PullImage(ctx context.Context, ref string, tag string) error {
	refStr := ref + ":" + tag
	reader, err := dClient.Client.ImagePull(ctx, refStr, image.PullOptions{})
	if err != nil {
		return err
	}
	_, err = io.Copy(os.Stdout, reader)
	if err != nil {
		return err
	}
	reader.Close()
	fmt.Printf("[Docker]: Pulled %s image from docker registry\n", refStr)
	return nil
}

func (dc *DockerContainer) GetName() string {
	return dc.ContainerName
}

func (dc *DockerContainer) GetID() string {
	return dc.ContainerID
}

func (dc *DockerContainer) GetTag() string {
	return dc.Tag
}

// TODO: Separate container creation from Run to that run can just run an existing container
// by being provided with an ID

// Runs and existing container, if it doesnt then it creates the container from
// dockerConfig and then runs the container afterwards
func (dc *DockerContainer) Run(dctx context.Context, imageName string, imageTag string) <-chan error {

	// check architecture and OS

	fmt.Printf("%s: checking for existing container before running...\n", dc.ContainerName)
	errChan := make(chan error, 1)

	// This getter only searches for a <zensearch> specific container
	// eg: zensearch-rabbitmq, zensearch-mysql, zensearch-selenium-*

	c, exists := dc.GetContainer(dctx)
	fmt.Printf("%s: Checking if container exists\n", dc.ContainerName)
	if exists {
		fmt.Printf("Docker: %s container already exist\n", dc.ContainerName)
		err := dc.Start(dctx, c.ID)
		// can be nil
		errChan <- err
		if err != nil {
			fmt.Printf("%s: unable to start from existing container...\n", dc.ContainerName)
			return errChan
		}
		go dc.listenContainerState(dctx)
		fmt.Printf("%s: container exposed ports -> %+v\n", dc.ContainerName, dc.HostPorts)
		return errChan
	}

	fmt.Printf("%s: creating new container...\n", dc.ContainerName)
	err := dc.Create(dctx, imageName, imageTag)
	if err != nil {
		fmt.Printf("%s: Unable to create container\n", dc.ContainerName)
		errChan <- err
		return errChan
	}

	fmt.Printf("%s: starting container...\n", dc.ContainerName)

	if err := dc.Client.ContainerStart(dctx, dc.ContainerID, container.StartOptions{}); err != nil {
		fmt.Printf("%s: Unable to start container\n", dc.ContainerName)
		errChan <- err
		return errChan
	}

	// dont know when it is completely finished, need to set a timer for other
	// process that depends on rabbitmq

	go dc.listenContainerState(dctx)
	fmt.Printf("%s: container started!\n", dc.ContainerName)
	fmt.Printf("%s: container exposed ports -> %+v\n", dc.ContainerName, dc.HostPorts)

	errChan <- nil
	return errChan
}

// TODO how to start already existing container?
// which means a container not created programmatically in here
// but instead passing in an existing ContainerID in the user's docker container list
func (dc *DockerContainer) Start(dctx context.Context, containerID string) error {

	fmt.Printf("%s: starting container...\n", dc.ContainerName)
	if containerID != "" {
		dc.ContainerID = containerID
		fmt.Printf("%s: assigning container ID for container\n", dc.ContainerName)
	}
	if dc.ContainerID == "" {
		return fmt.Errorf("%s: ERROR current container does not have an associated ContainerID which means the container does not exist, instead run the Run() function to create and run a new container from an image\n", dc.ContainerName)
	}

	err := dc.Client.ContainerStart(dctx, dc.ContainerID, container.StartOptions{})
	if err != nil {
		fmt.Printf("%s: Unable to start the container", dc.ContainerName)
		return err
	}
	fmt.Printf("%s: container started!\n", dc.ContainerName)

	return nil
}

// TODO POINTER NOT BEING UPDATED WHEN USING EXISTING CONTAINER
func (dc *DockerContainer) Stop(dctx context.Context) error {
	if dc.ContainerID == "" {
		return fmt.Errorf("%s: ERROR there's nothing to stop because the container does not exist\n", dc.ContainerName)
	}
	fmt.Printf("%s: stopping container...\n", dc.ContainerName)
	err := dc.Client.ContainerStop(dctx, dc.ContainerID, container.StopOptions{Signal: "SIGKILL"})
	if err != nil {
		fmt.Printf("%s: ERROR %s", dc.ContainerName, err)
		return fmt.Errorf("Docker: ERROR Something went wrong, zensearch is unable to stop the container %s with ID of %s\n", dc.ContainerID[:8], dc.ContainerName)
	}
	fmt.Printf("%s: Successfully stopped with ID starting with %s\n", dc.ContainerName, dc.ContainerID[:8])
	return nil
}

// Creates a new container and updates the cc's ContainerID field is successful
// else will panic dont use separately from Run() because port mapping is only initialized
// on container startup and not on creation... i dont know why
func (dc *DockerContainer) Create(dctx context.Context, imageName string, tag string) error {
	fmt.Printf("%s: creating container...\n", dc.ContainerName)
	imageNameWithTag := imageName + ":" + tag
	fmt.Printf("%s: applying ports\n", dc.ContainerName)
	hostPorts := map[nat.Port][]nat.PortBinding{}
	for _, hostPort := range dc.HostPorts {
		_, ok := hostPorts[nat.Port(fmt.Sprintf("%s/tcp", hostPort))]
		if !ok {
			ports := hostPorts[nat.Port(fmt.Sprintf("%s/tcp", hostPort))]
			hostPorts[nat.Port(fmt.Sprintf("%s/tcp", hostPort))] = append(ports, nat.PortBinding{HostIP: "0.0.0.0", HostPort: hostPort})
			ports = hostPorts[nat.Port(fmt.Sprintf("%s/tcp", hostPort))]
		}
	}
	containerPorts := map[nat.Port]struct{}{}
	for _, contPort := range dc.ContainerPorts {
		_, ok := hostPorts[nat.Port(fmt.Sprintf("%s/tcp", contPort[0]))]
		if !ok {
			containerPorts[nat.Port(fmt.Sprintf("%s/tcp", contPort[0]))] = struct{}{}
		}
	}

	fmt.Printf("Docker: creating container from %s image as %s \n", imageNameWithTag, dc.ContainerName)

	// TODO ERROR from here for some reason
	resp, err := dc.Client.ContainerCreate(dctx, &container.Config{
		Image: imageNameWithTag,
		// attaching container to process exec is on `-it`
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		ExposedPorts: containerPorts,
		Env:          dc.Env,
	},
		&container.HostConfig{
			ShmSize: dc.ShmSize,
			Binds: []string{
				"/var/run/docker.sock:/var/run/docker.sock",
			},
			PortBindings: hostPorts}, nil, nil, dc.ContainerName)

	if err != nil {
		fmt.Print("Invalid reference format error from here\n")
		return err
	}
	fmt.Printf("Docker: %s's container ID %s\n", dc.ContainerName, resp.ID)
	dc.ContainerID = resp.ID
	return nil
}

// TODO figure out what to do with this
func (dc *DockerContainer) listenContainerState(dctx context.Context) {
	fmt.Printf("\n%s: waiting for container status...\n", dc.ContainerName)
	statusCh, errCh := dc.Client.ContainerWait(dctx, dc.ContainerID, container.WaitConditionNotRunning)
	// Listening to stdout of container
	go func() {
		out, err := dc.Client.ContainerLogs(dctx, dc.ContainerID, container.LogsOptions{ShowStdout: true, ShowStderr: true})

		if err != nil {
			fmt.Println(err.Error())
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
		fmt.Printf("%s: closing container\n", dc.ContainerName)
		fmt.Printf("%s: cause for closing container: %s\n", dc.ContainerName, err.Error())
		return
	case s := <-statusCh:
		fmt.Printf("Container %s status:\n", dc.ContainerName)
		if s.Error != nil {
			fmt.Println(s.Error)
			return
		}
		fmt.Printf("Docker: %s container closed gracefully\n", dc.ContainerName)
	}
}

// should check if a rabbitmq container already exists
func (dc *DockerContainer) GetContainer(dctx context.Context) (container.Summary, bool) {
	filter := filters.NewArgs()
	filter.Add("name", dc.ContainerName)
	fmt.Println("WHAT ", dc.ContainerName)

	containers, err := dc.Client.ContainerList(dctx, container.ListOptions{Size: false, Filters: filter, All: true})
	if err != nil {
		fmt.Println("Docker: ERROR Unable to get list of containers")
		return container.Summary{}, false
	}
	if len(containers) == 0 {
		fmt.Printf("Docker: container %s does not exist\n", dc.ContainerName)
		return container.Summary{}, false
	}
	fmt.Println(containers[0].ID)
	return containers[0], true

}
