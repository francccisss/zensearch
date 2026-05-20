package main

import (
	"bufio"
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

var user = os.Getenv("USER")
var DOCKER_DESKTOP_HOST = fmt.Sprintf("unix:///home/%s/.docker/desktop/docker.sock", user)

const STANDALONE_DOCKER_HOST = "unix:///var/run/docker.sock"

// host is the current socket used by dockerd to communicate with the client
// make it so that if an error is in relation to a unreachable host(docker engine)
// then we can check the other host url since both Docker Desktop and standalone Docker Engine
// uses 2 different URL Socket.

func NewDockerManager() (DockerManager, error) {
	cl, err := client.NewClientWithOpts(client.FromEnv, client.WithHost(DOCKER_DESKTOP_HOST), client.WithAPIVersionNegotiation())
	if err != nil {
		return DockerManager{}, err
	}

	dcmap := make(map[ContainerName]*DockerContainer, 2)
	return DockerManager{
		Client:     cl,
		Containers: &dcmap,
		Images:     make(map[string]*DockerImage, 2),
		Host:       cl.DaemonHost(),
	}, nil
}

func (dm *DockerManager) NewDockerContainer(contConfig DockerContainerConfig) {

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
	GetContainer(dctx context.Context) (container.Summary, bool)
	listenContainerState(dctx context.Context)
	RemoveContainer(ctx context.Context, containerName ContainerName)
}

// Run() pulls an image from the docker registry given the container configuration
// created with NewDockerContainer,

// even though that the docker engine alreaady returns the existing image its better to
// check first if the image already exists before calling ImagePull from docker registry
func (dm *DockerManager) PullImage(ctx context.Context, ref string) error {

	filterArgs := filters.NewArgs(filters.KeyValuePair{Key: "reference", Value: ref})

	fmt.Printf("[DOCKER]: Checking if an image already exists for %s\n", ref)
	list, err := dm.GetImageList(ctx, image.ListOptions{Filters: filterArgs})
	if err != nil {
		return err
	}
	if len(list) > 0 {
		fmt.Printf("[DOCKER]: Image %s already exists\n", ref)
		fmt.Println(list[0])
		return nil
	}

	reader, err := dm.Client.ImagePull(ctx, ref, image.PullOptions{})
	if err != nil {
		return err
	}
	stdoutReader, stdouWriter := io.Pipe()
	scanner := bufio.NewScanner(stdoutReader)
	go func() {
		for scanner.Scan() {
			fmt.Printf("[DOCKER]: IMAGE PULL - '%s'\n", scanner.Text())
		}
	}()
	_, err = io.Copy(stdouWriter, reader)
	if err != nil {
		return err
	}
	reader.Close()
	return nil
}

func (dm *DockerManager) Run(dctx context.Context, containerName ContainerName, ref string) <-chan error {

	// check architecture and OS
	errChan := make(chan error, 1)

	dockerContainer, ok := (*dm.Containers)[containerName]
	if !ok {
		errChan <- fmt.Errorf("[DOCKER]: ERROR - '%s Container does not exist'", containerName)
		return errChan
	}
	err := dm.Create(dctx, containerName, ref)
	if err != nil {
		fmt.Printf("[DOCKER]: Unable to create %s container\n", containerName)
		errChan <- err
		return errChan
	}

	fmt.Printf("[DOCKER]: starting %s container\n", containerName)

	if err := dm.Client.ContainerStart(dctx, dockerContainer.ContainerID, container.StartOptions{}); err != nil {
		fmt.Printf("[DOCKER]: Unable to start %s container\n", containerName)
		errChan <- err
		return errChan
	}

	// dont know when it is completely finished, need to set a timer for other
	// process that depends on rabbitmq

	// go dm.listenContainerState(dctx, containerName)
	fmt.Printf("[DOCKER]: %s Container started!\n", containerName)
	fmt.Printf("[DOCKER]: %s Container exposed ports -> %+v\n", containerName, dockerContainer.HostPorts)

	errChan <- nil
	return errChan
}

// TODO how to start already existing container?
// which means a container not created programmatically in here
// but instead passing in an existing ContainerID in the user's docker container list
func (dm *DockerManager) Start(dctx context.Context, containerID string) error {

	fmt.Printf("[DOCKER]: starting container %s \n", containerID)

	if containerID == "" {
		return fmt.Errorf("[DOCKER]: ERROR current container does not have an associated ContainerID which means the container does not exist, instead run the Run() function to create and run a new container from an image\n")
	}
	err := dm.Client.ContainerStart(dctx, containerID, container.StartOptions{})
	if err != nil {
		return err
	}
	fmt.Printf("[DOCKER]: Container %s started!\n", containerID[:8])
	return nil
}

// TODO POINTER NOT BEING UPDATED WHEN USING EXISTING CONTAINER
func (dm *DockerManager) Stop(dctx context.Context, containerName ContainerName) error {
	dockerContainer, ok := (*dm.Containers)[containerName]
	if !ok {
		return fmt.Errorf("[DOCKER]: ERROR - '%s Container does not exist'", containerName)
	}

	fmt.Printf("[DOCKER]: Stopping %s container...\n", containerName)
	err := dm.Client.ContainerStop(dctx, dockerContainer.ContainerID, container.StopOptions{Signal: "SIGKILL"})
	if err != nil {
		fmt.Printf("[DOCKER]: ERROR - %s container - '%s'\n", containerName, err)

		return fmt.Errorf("[DOCKER]: ERROR Something went wrong, zensearch is unable to stop the container %s with ID of %s\n", dockerContainer.ContainerID[:8], containerName)
	}
	fmt.Printf("[DOCKER]: Successfully stopped a container with an ID starting with %s\n", dockerContainer.ContainerID[:8])
	return nil
}

// Before creating a container, it checks for an existing continer first and uses that instead and updates the ID
// of the DockerContainer struct type.
func (dm *DockerManager) Create(dctx context.Context, containerName ContainerName, ref string) error {

	fmt.Printf("[DOCKER]: Checking if %s container exists first\n", containerName)
	dockerContainer, ok := (*dm.Containers)[containerName]
	if !ok {
		return fmt.Errorf("[DOCKER]: DOCKER MANAGER ERROR - '%s Container Struct does not exist'", containerName)
	}

	sum, exists := dm.GetContainer(dctx, dockerContainer.ContainerName)
	if exists {
		dockerContainer.ContainerID = sum.ID
		fmt.Printf("[DOCKER]: %s container already exist\n", dockerContainer.ContainerID)
		return nil
	}

	fmt.Printf("[DOCKER]: Creating %s container...\n", containerName)
	fmt.Printf("[DOCKER]: Applying ports for %s\n", containerName)
	hostPorts := make(map[nat.Port][]nat.PortBinding)
	for _, hostPort := range dockerContainer.HostPorts {
		hostPorts[hostPort] = append(hostPorts[hostPort], nat.PortBinding{HostIP: "0.0.0.0", HostPort: string(hostPort)})
	}

	fmt.Printf("[DOCKER]: Creating container from %s image as %s \n", ref, containerName)

	// TODO dont include the env and shmsize for rabbitmq
	resp, err := dm.Client.ContainerCreate(dctx, &container.Config{
		Image: ref,
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
	fmt.Printf("[DOCKER]: %s's container ID %s\n", containerName, resp.ID)
	dockerContainer.ContainerID = resp.ID
	return nil
}

func (dm *DockerManager) listenContainerState(dctx context.Context, containerName ContainerName) {
	dockerContainer, ok := (*dm.Containers)[containerName]
	if !ok {
		panic(fmt.Sprintf("[DOCKER]: ERROR - '%s Container does not exist'", containerName))
	}

	fmt.Printf("[DOCKER]: Waiting for %s's status\n", containerName)
	statusCh, errCh := dm.Client.ContainerWait(dctx, dockerContainer.ContainerID, container.WaitConditionNotRunning)
	// reading from go routine errors
	IOErrCh := make(chan error, 1)
	// Listening to stdout of container
	out, err := dm.Client.ContainerLogs(dctx, dockerContainer.ContainerID, container.LogsOptions{ShowStdout: true, ShowStderr: true, Follow: true})
	if err != nil {
		fmt.Println()
		panic(fmt.Sprintf("[DOCKER]: ERROR - 'Unable to capture stream from container - CAUSE: %s'", err.Error()))
	}
	defer func() {
		if closeErr := out.Close(); closeErr != nil {
			IOErrCh <- closeErr
		}
	}()

	stdoutReader, stdoutWriter := io.Pipe()
	stderrReader, stderrWriter := io.Pipe()

	go func() {
		dstout := bufio.NewScanner(stdoutReader)
		for dstout.Scan() {
			select {
			case <-dctx.Done():
				fmt.Println("[DOCKER]: Closing Stdout tail")
				return
			default:
				fmt.Printf("[DOCKER]: STATUS - %s, '%s'\n", containerName, dstout.Text())
			}

		}
	}()

	go func() {
		dsterr := bufio.NewScanner(stderrReader)
		for dsterr.Scan() {
			select {
			case <-dctx.Done():
				fmt.Println("[DOCKER]: Closing Error tail")
				return
			default:
				fmt.Printf("[DOCKER]: STATUS - %s, '%s'\n", containerName, dsterr.Text())
			}
		}
	}()

	go func() {
		_, err = stdcopy.StdCopy(stdoutWriter, stderrWriter, out)
		if err != nil {
			fmt.Println(err.Error())
			IOErrCh <- err
		}
	}()

	select {
	case <-dctx.Done():
		if dctx.Err().Error() == context.Canceled.Error() {
			fmt.Printf("[DOCKER]: CONTAINER ERROR - 'Context' - CAUSE: %s\n", dctx.Err().Error())
		}
		return
	case closeErr := <-IOErrCh:
		panic(fmt.Sprintf("[DOCKER]: CONTAINER ERROR - 'Error occured while closing' - CAUSE: %s\n", closeErr))
	case err := <-errCh:
		panic(fmt.Sprintf("[DOCKER]: CONTAINER ERROR - 'Error occured while listening from container stream - CAUSE: '%s'\n", err.Error()))
	case s := <-statusCh:
		if s.Error != nil {
			panic(fmt.Sprintf("[DOCKER]: CONTAINER ERROR - 'Error occured while listening from container stream - CAUSE: '%s'\n", s.Error.Message))
		}
		fmt.Printf("[DOCKER]: %s container closed gracefully - STATUS:%+v \n", containerName, s)
	}
}

func (dm *DockerManager) GetImageList(dctx context.Context, options image.ListOptions) ([]image.Summary, error) {
	list, err := dm.Client.ImageList(dctx, options)
	if err != nil {
		return []image.Summary{}, err
	}
	return list, nil
}

// should check if a rabbitmq container already exists
func (dm *DockerManager) GetContainer(dctx context.Context, containerName ContainerName) (container.Summary, bool) {
	filter := filters.NewArgs()
	filter.Add("name", string(containerName))
	containers, err := dm.Client.ContainerList(dctx, container.ListOptions{Size: false, Filters: filter, All: true})
	if err != nil {
		fmt.Println("[DOCKER]: ERROR Unable to get list of containers")
		return container.Summary{}, false
	}
	if len(containers) == 0 {
		fmt.Printf("[DOCKER]: container %s does not exist\n", containerName)
		return container.Summary{}, false
	}
	fmt.Printf("[DOCKER]: %s Container exists \n", containerName)
	return containers[0], true

}

func (dm *DockerManager) RemoveContainer(ctx context.Context, containerName ContainerName) error {

	dockerContainer, ok := (*dm.Containers)[containerName]
	if !ok {
		panic(fmt.Sprintf("[DOCKER]: ERROR - '%s Container does not exist'", containerName))
	}
	err := dm.Stop(ctx, containerName)
	if err != nil {
		return err
	}
	err = dm.Client.ContainerRemove(ctx, dockerContainer.ContainerID, container.RemoveOptions{Force: true})
	if err != nil {
		return err
	}
	return nil
}
