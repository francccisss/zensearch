package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
)

var wg sync.WaitGroup

// const dockerArgs = "-it --rm --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:4.0-management"

func stringToArr(str string) []string {
	tmp := []string{}
	startP := 0
	for i := range str {
		if str[i] == ' ' {
			tmp = append(tmp, str[startP:i])
			startP = i + 1
		}
	}
	tmp = append(tmp, str[startP:])
	return tmp
}

// TODO contain in docker/containers.go
// create separate image and container initialization
// for rabbitmq and selenium containers

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
	Run()
	Start()
	// Restart()
	Create(ctx context.Context)
	Stop()
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

// Pulls the image from the registry then creates the container
func (cc *ClientContainer) Run(ctx context.Context, imageName string, tag string) {

	imageNameWithTag := imageName + ":" + tag
	fmt.Printf("Docker: pulling %s image...\n", imageNameWithTag)
	reader, err := cc.Client.ImagePull(ctx, imageName+":"+tag, image.PullOptions{})
	if err != nil {
		log.Panic(err)
	}

	io.Copy(os.Stdout, reader)
	defer reader.Close()

	cc.Create(ctx, imageName, tag)
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

// Creates a new container and updates the clientContainer's ContainerID field is successful
// else will panic
func (cc *ClientContainer) Create(ctx context.Context, imageName string, tag string) {

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

func TestDockerRabbitmq(t *testing.T) {

	ctx := context.Background()
	fmt.Printf("Docker: connecting client to docker daemon...\n")
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Panic(err.Error())
	}
	clientContainer := ClientContainer{
		Client:         cli,
		ContainerName:  "zensearch-cli-rabbitmq",
		HostPorts:      HostPorts{"5672", "15672"},
		ContainerPorts: ContainerPorts{{"5672", "5672"}, {"15672", "15672"}}}
	defer cli.Close()

	_, err = clientContainer.getContainer(ctx)

	// if container does not exist create make sure to pull needed image
	// then create
	if err != nil {
		fmt.Printf("Docker: %s\n", err.Error())
		clientContainer.Run(ctx, "rabbitmq", "4.0-management")
		// clientContainer.Create(ctx, "rabbitmq", "4.0-management")
	}

}

func TestCommandExec(t *testing.T) {
	for _, build := range buildCmd {

		cmd := exec.Command(build[1], build[2:]...)
		stdErr, err := cmd.StderrPipe()
		// stdOut, err := cmd.StdoutPipe()
		err = cmd.Start()
		if err != nil {
			fmt.Println("Error: cannot run command")
			t.Fail()
			switch e := err.(type) {
			case *exec.Error:
				fmt.Println("failed executing:", err)
				break
			case *exec.ExitError:
				readStdErr, err := io.ReadAll(stdErr)
				if err != nil {
					fmt.Println(err.Error())
				}
				fmt.Println("command exit rc =", e.ExitCode())
				fmt.Printf("%s> %s\n", build[0], string(readStdErr))
				panic(err)
			default:
				panic(err)
			}
			errArr = append(errArr, []string{build[0], err.Error()})
		}
		fmt.Printf("%s: building %s service...\n", build[0], build[0])
		if build[0] == "database" {
			fmt.Println("NOTICE FOR DATABASE SERVICE: make sure you have sqlite3 installed on your system!")
		}
		// readStdOut, err := io.ReadAll(stdOut)
		if err != nil {
			fmt.Println(err.Error())
		}
		cmd.Wait()
		switch build[1] {
		case "go":
			fmt.Printf("%s: build successful\n", build[0])
			break
		case "npm":
			fmt.Printf("%s: installed node dependencies\n", build[0])
			break
		}
	}

}
