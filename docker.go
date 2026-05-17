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

type ContainerNameOption int

type ContainerName string

type DockerContainerConfig struct {
	ImageName string
	Tag       string
	Name      ContainerName
	HostPorts
	ContainerPorts nat.PortSet
	ShmSize        int64
	Env            []string
}

type HostPorts []nat.Port

// TODO write context purpose of dtx

type DockerContainer struct {
	*client.Client
	HostPorts
	ContainerPorts nat.PortSet
	ContainerName
	ContainerID string
	ShmSize     int64
	Env         []string
	Tag         string
}

type DockerImage struct {
	ImageName string
	Tag       string
}

type DockerManager struct {
	Client     *client.Client
	Containers *map[ContainerName]*DockerContainer
	Images     map[string]*DockerImage
	Host       string
}

// host is the current socket used by dockerd to communicate with the client
func NewDockerManager(Host string) (DockerManager, error) {
	cl, err := client.NewClientWithOpts(client.FromEnv, client.WithHost(Host))
	if err != nil {
		return DockerManager{}, nil
	}
	dcmap := make(map[ContainerName]*DockerContainer, 2)
	return DockerManager{
		Client:     cl,
		Containers: &dcmap,
		Images:     make(map[string]*DockerImage, 2),
		Host:       Host,
	}, nil
}

func (dm *DockerManager) NewDockerContainer(contConfig DockerContainerConfig) {
	fmt.Printf("[New Docker Container NOTICE]: Creating new docker container for %s\n", contConfig.Name)

	newDockerContainer := &DockerContainer{
		Client:         dm.Client,
		ContainerName:  contConfig.Name,
		HostPorts:      contConfig.HostPorts,
		ContainerPorts: contConfig.ContainerPorts,
		ShmSize:        contConfig.ShmSize,
		Env:            contConfig.Env,
		Tag:            contConfig.Tag,
		ContainerID:    "",
	}
	(*dm.Containers)[newDockerContainer.ContainerName] = newDockerContainer
}

type ContainerManager interface {
	Run(dctx context.Context, imageName string, imageTag string) <-chan error
	Create(dctx context.Context, imageName string, tag string) error
	Start(dctx context.Context, containerID string) error
	Stop(dctx context.Context) error
	GetContainerID(dctx context.Context) (container.Summary, bool)
	listenContainerState(dctx context.Context)
	RemoveContainer(ctx context.Context, containerName ContainerName)
}

// Run() pulls an image from the docker registry given the container configuration
// created with NewDockerContainer,

// even though that the docker engine alreaady returns the existing image its better to
// check first if the image already exists before calling ImagePull from docker registry

func (dm *DockerManager) PullImage(ctx context.Context, ref string, tag string) error {
	refStr := ref + ":" + tag
	reader, err := dm.Client.ImagePull(ctx, refStr, image.PullOptions{})
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

func (dm *DockerManager) Run(dctx context.Context, containerName ContainerName, imageName string, tag string) <-chan error {

	// check architecture and OS
	fmt.Printf("[Docker]: checking for existing container before running...\n")
	errChan := make(chan error, 1)

	dockerContainer, ok := (*dm.Containers)[containerName]
	if !ok {
		errChan <- fmt.Errorf("[Docker]: ERROR - '%s Container does not exist'", containerName)
		return errChan
	}
	err := dm.Create(dctx, containerName, imageName, tag)
	if err != nil {
		fmt.Printf("%s: Unable to create container\n", containerName)
		errChan <- err
		return errChan
	}

	fmt.Printf("%s: starting container...\n", containerName)

	if err := dm.Client.ContainerStart(dctx, dockerContainer.ContainerID, container.StartOptions{}); err != nil {
		fmt.Printf("%s: Unable to start container\n", containerName)
		errChan <- err
		return errChan
	}

	// dont know when it is completely finished, need to set a timer for other
	// process that depends on rabbitmq

	go dm.listenContainerState(dctx, containerName)
	fmt.Printf("%s: container started!\n", containerName)
	fmt.Printf("%s: container exposed ports -> %+v\n", containerName, dockerContainer.HostPorts)

	errChan <- nil
	return errChan
}

// TODO how to start already existing container?
// which means a container not created programmatically in here
// but instead passing in an existing ContainerID in the user's docker container list
func (dm *DockerManager) Start(dctx context.Context, containerID string) error {

	fmt.Printf("[Docker]: starting container %s \n", containerID)

	if containerID == "" {
		return fmt.Errorf("[Docker]: ERROR current container does not have an associated ContainerID which means the container does not exist, instead run the Run() function to create and run a new container from an image\n")
	}
	err := dm.Client.ContainerStart(dctx, containerID, container.StartOptions{})
	if err != nil {
		return err
	}
	fmt.Printf("[Docker]: Container %s started!\n", containerID[:8])
	return nil
}

// TODO POINTER NOT BEING UPDATED WHEN USING EXISTING CONTAINER
func (dm *DockerManager) Stop(dctx context.Context, containerName ContainerName) error {
	dockerContainer, ok := (*dm.Containers)[containerName]
	if !ok {
		return fmt.Errorf("[Docker]: ERROR - '%s Container does not exist'", containerName)
	}

	fmt.Printf("[Docker]: Stopping %s container...\n", containerName)
	err := dm.Client.ContainerStop(dctx, dockerContainer.ContainerID, container.StopOptions{Signal: "SIGKILL"})
	if err != nil {
		fmt.Printf("[Docker]: ERROR - %s container - '%s'\n", containerName, err)

		return fmt.Errorf("Docker: ERROR Something went wrong, zensearch is unable to stop the container %s with ID of %s\n", dockerContainer.ContainerID[:8], containerName)
	}
	fmt.Printf("%s: Successfully stopped with ID starting with %s\n", containerName, dockerContainer.ContainerID[:8])
	return nil
}

func (dm *DockerManager) Create(dctx context.Context, containerName ContainerName, imageName string, tag string) error {

	fmt.Printf("[Docker]: Checking if %s container exists first\n", containerName)
	dockerContainer, ok := (*dm.Containers)[containerName]
	if !ok {
		return fmt.Errorf("[Docker]: ERROR - '%s Container does not exist'", containerName)
	}

	sum, exists := dm.GetContainerID(dctx, dockerContainer.ContainerName)
	if exists {
		dockerContainer.ContainerID = sum.ID
		fmt.Printf("[Docker]: %s container already exist\n", dockerContainer.ContainerID)
		return nil
	}

	fmt.Printf("[Docker]: Creating %s container...\n", containerName)
	imageNameWithTag := imageName + ":" + tag
	fmt.Printf("[Docker]: Applying ports for %s\n", containerName)
	hostPorts := make(map[nat.Port][]nat.PortBinding)
	for _, hostPort := range dockerContainer.HostPorts {
		hostPorts[hostPort] = append(hostPorts[hostPort], nat.PortBinding{HostIP: "0.0.0.0", HostPort: string(hostPort)})
	}

	fmt.Printf("Docker: creating container from %s image as %s \n", imageNameWithTag, containerName)

	// TODO dont include the env and shmsize for rabbitmq
	resp, err := dm.Client.ContainerCreate(dctx, &container.Config{
		Image: imageNameWithTag,
		// attaching container to process exec is on `-it`
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		ExposedPorts: dockerContainer.ContainerPorts,
		Env:          dockerContainer.Env,
	},
		&container.HostConfig{
			ShmSize: dockerContainer.ShmSize,
			// TODO PASS IN MANAGER HOST SOCKET
			Binds: []string{
				"/var/run/docker.sock:/var/run/docker.sock",
			},
			PortBindings: hostPorts}, nil, nil, string(containerName))

	if err != nil {
		fmt.Print("Invalid reference format error from here\n")
		return err
	}
	fmt.Printf("Docker: %s's container ID %s\n", containerName, resp.ID)
	dockerContainer.ContainerID = resp.ID
	return nil
}

func (dm *DockerManager) listenContainerState(dctx context.Context, containerName ContainerName) {
	dockerContainer, ok := (*dm.Containers)[containerName]
	if !ok {
		fmt.Printf("[Docker]: ERROR - '%s Container does not exist'", containerName)
		return
	}

	fmt.Printf("\n%s: waiting for container status...\n", containerName)
	statusCh, errCh := dm.Client.ContainerWait(dctx, dockerContainer.ContainerID, container.WaitConditionNotRunning)
	// Listening to stdout of container
	go func() {
		out, err := dm.Client.ContainerLogs(dctx, dockerContainer.ContainerID, container.LogsOptions{ShowStdout: true, ShowStderr: true})

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
		fmt.Printf("[Docker]: %s - Closing container\n", containerName)
		fmt.Printf("[Docker]: %s - Cause for closing container: '%s'\n", containerName, err.Error())
		return
	case s := <-statusCh:
		fmt.Printf("Container %s status:\n", containerName)
		if s.Error != nil {
			fmt.Println(s.Error)
			return
		}
		fmt.Printf("Docker: %s container closed gracefully\n", containerName)
	}
}

// should check if a rabbitmq container already exists
func (dm *DockerManager) GetContainerID(dctx context.Context, containerName ContainerName) (container.Summary, bool) {
	filter := filters.NewArgs()
	filter.Add("name", string(containerName))
	containers, err := dm.Client.ContainerList(dctx, container.ListOptions{Size: false, Filters: filter, All: true})
	if err != nil {
		fmt.Println("Docker: ERROR Unable to get list of containers")
		return container.Summary{}, false
	}
	if len(containers) == 0 {
		fmt.Printf("Docker: container %s does not exist\n", containerName)
		return container.Summary{}, false
	}
	return containers[0], true

}

func (dm *DockerManager) RemoveContainer(ctx context.Context, containerName ContainerName) error {

	dockerContainer, ok := (*dm.Containers)[containerName]
	err := dm.Stop(ctx, containerName)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("[Docker]: ERROR - '%s Container does not exist'", containerName)
	}
	err = dm.Client.ContainerRemove(ctx, dockerContainer.ContainerID, container.RemoveOptions{Force: true})
	if err != nil {
		return err
	}
	return nil
}
