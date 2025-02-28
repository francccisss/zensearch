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
)

var wg sync.WaitGroup

// const dockerArgs = "-it --rm --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:4.0-management"
const dockerArgs = "-it --rm -p 5672:5672 -p 15672:15672 "

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

func TestDockerRabbitmq(t *testing.T) {

	fmt.Println("Docker: Starting docker...")
	fmt.Printf("Docker: Args -> %s\n", dockerArgs)

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Panic(err.Error())
	}
	defer cli.Close()

	var currentContainerID string
	imageName := "rabbitmq"
	setContainerName := "zensearch-rabbitmq"
	filter := filters.NewArgs()
	filter.Add("name", imageName)

	containers, err := cli.ContainerList(ctx, container.ListOptions{Size: false, Filters: filter, All: true})
	if err != nil {
		fmt.Println("Unable to get list of containers")
		panic(err)
	}

	if len(containers) == 0 {

		fmt.Printf("Docker: pulling latest %s image...\n", imageName)
		reader, err := cli.ImagePull(ctx, imageName, image.PullOptions{})
		if err != nil {
			log.Panic(err)
		}

		io.Copy(os.Stdout, reader)
		defer reader.Close()

		// grabs latest version of rabbitmq
		fmt.Printf("Docker: creating container from %s image as %s \n", imageName, setContainerName)
		resp, err := cli.ContainerCreate(ctx, &container.Config{
			Image:        "rabbitmq",
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			Cmd:          stringToArr(dockerArgs),
		}, nil, nil, nil, imageName)

		fmt.Printf("Docker: %s's container ID %s\n", setContainerName, resp.ID)
		currentContainerID = resp.ID
	} else {
		currentContainerID = containers[0].ID
		fmt.Printf("Docker: container already exists from %s image as %s \n", imageName, containers[0].Names)
		fmt.Printf("Docker: %s's container ID %s\n", setContainerName, currentContainerID)
	}
	fmt.Printf("Docker: starting %s container...", setContainerName)
	if err := cli.ContainerStart(ctx, currentContainerID, container.StartOptions{}); err != nil {
		fmt.Println("Unable to start rabbitmq container")
		panic(err)
	}

	statusCh, errCh := cli.ContainerWait(ctx, currentContainerID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		fmt.Println("Container state not positive")
		panic(err)
	case s := <-statusCh:
		fmt.Println("Container status:")
		fmt.Println(s.Error.Message)
	}

	out, err := cli.ContainerLogs(ctx, currentContainerID, container.LogsOptions{ShowStdout: true, ShowStderr: true})
	stdcopy.StdCopy(os.Stdout, os.Stderr, out)
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
