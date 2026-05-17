package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

type StdError struct {
	src   string
	value string
}

func NewError(src string) StdError {
	return StdError{src: src}
}

func (se StdError) Error() string {
	return fmt.Sprintf("[ZENSEARCH]: ERROR FROM - %s - '%s'", se.src, se.value)
}

func (se *StdError) addError(value string) {
	se.value = value
}

// TODO make sure docker services are running first
// use go routines and wait for state changes
// need to start containerized services first eg: selenium/rabbitmq
// tail each docker container service until reading from
// stdin returns a <name> started successfully

// TODO: make it so that if users can use either docker desktop or just dockerd
// so that they configure the location of the host to their dockerd socket path

func startServices(pctx context.Context, commands [][]string) {
	dockerMan, err := NewDockerManager()
	if err != nil {
		log.Fatalf("[Start Service ERROR]: '%s'", err)
	}
	fmt.Println("[ZENSEARCH]: Starting services...")
	ctx, cancel := context.WithCancel(pctx)
	defer cancel()

	var wg sync.WaitGroup

	fmt.Println("[DOCKER]: Starting up Docker Services")
	for _, contConfig := range dockerContainerConf {
		wg.Go(func() {
			runningDockerService(ctx, &dockerMan, contConfig)
		})
	}

	wg.Wait()

	fmt.Println("[DOCKER]: Starting up Express & DB Services")
	for _, command := range commands {
		wg.Go(func() {
			cmd := exec.Command(command[1], command[2:]...)
			runningService(ctx, cmd, command[0])
		})
	}
	fmt.Println("[ZENSEARCH]: services started")
}

func runningDockerService(ctx context.Context, dockerMan *DockerManager, contConfig DockerContainerConfig) {
	newErr := NewError(string(contConfig.Name))

	dockerMan.NewDockerContainer(contConfig)

	err := dockerMan.PullImage(ctx, contConfig.ImageName, contConfig.Tag)
	if err != nil {
		newErr.addError(err.Error())
		panic(newErr.Error)
	}
	// cancellation for specific service
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*120)
	defer cancel()
	go func() {
		<-ctx.Done()
		if err := dockerMan.Stop(context.Background(), contConfig.Name); err != nil {
			newErr.addError(err.Error())
			fmt.Println(newErr.Error())
			return
		}
		fmt.Printf("[DOCKER]: %s container stopped\n", contConfig.Name)
	}()

	select {
	case err := <-dockerMan.Run(ctx, contConfig.Name, contConfig.ImageName, contConfig.Tag):
		if err != nil {
			newErr.addError(err.Error())
			panic(newErr.Error)
		}
		fmt.Printf("[DOCKER]: %s Container Successfuly started!\n", contConfig.Name)
	case <-timeoutCtx.Done():
		if timeoutCtx.Err() == context.DeadlineExceeded {
			newErr.addError(timeoutCtx.Err().Error())
			fmt.Printf("[DOCKER]: Failed to start %s!\n", contConfig.Name)
			fmt.Printf("[DOCKER]: %s Container timedout\n", contConfig.Name)
			panic(newErr.Error)
		}
	}

}
func runningService(ctx context.Context, cmd *exec.Cmd, cmdName string) {
	newStdErr := NewError(cmdName)
	stdout, err := cmd.StdoutPipe()
	stderr, err := cmd.StderrPipe()

	if err != nil {
		fmt.Printf("[ZENSEARCH]: ERROR unable to set up stdout for process %s\n", cmdName)
		newStdErr.addError(err.Error())
		panic(newStdErr.Error())
	}
	err = cmd.Start()
	if err != nil {
		fmt.Printf("[ZENSEARCH]: ERROR unable to start process %s\n", cmdName)
		newStdErr.addError(err.Error())
		panic(newStdErr.Error())
	}
	// for handling stderr
	go func() {
		readErrors, err := io.ReadAll(stderr)
		if err != nil {
			panic(newStdErr.Error())
		}
		newStdErr.addError(string(readErrors))
		panic(newStdErr.Error())
	}()

	io.Copy(os.Stdout, stdout)
	// use stderr to check the state of the process
	err = cmd.Wait()
	if err != nil {
		newStdErr.addError(err.Error())
		panic(newStdErr.Error())
	}

	<-ctx.Done()
	fmt.Printf("%s: Shut down\n", cmdName)

	if cmd.Process != nil {
		fmt.Printf("[ZENSEARCH]: %s - shutting down process...\n", cmdName)
		_ = cmd.Process.Signal(syscall.SIGTERM)
		fmt.Printf("[ZENSEARCH]: %s - closed\n", cmdName)
	}
}

func runCommands(commands [][]string) {
	newError := NewError("RUN COMMANDS")

	for _, command := range commands {
		cmd := exec.Command(command[1], command[2:]...)
		stdErr, _ := cmd.StderrPipe()
		stdOut, _ := cmd.StdoutPipe()
		err := cmd.Start()
		io.Copy(os.Stdout, stdOut)
		if err != nil {
			newError.addError(err.Error())
			fmt.Println("Error: cannot run command")
			switch e := err.(type) {
			case *exec.Error:
			case *exec.ExitError:
				readStdErr, err := io.ReadAll(stdErr)
				if err != nil {
					fmt.Println(err.Error())
				}
				fmt.Println("command exit rc =", e.ExitCode())
				fmt.Printf("%s> %s\n", command[0], string(readStdErr))
			}
			panic(newError.Error)
		}

		for _, str := range command {
			switch str {
			case "install":
				fmt.Printf("%s: installing dependencies for %s service...\n", command[0], command[0])
			case "build":
				fmt.Printf("%s: building %s service...\n", command[0], command[0])
			}
		}
		cmd.Wait()
		for _, str := range command {
			switch str {
			case "install":
				fmt.Printf("%s: dependencies successfully installed\n", command[0])
			case "build":
				fmt.Printf("%s: build successful\n", command[0])
			}
		}
	}

}

func help() {
	fmt.Printf(`
Welcome to zensearch cli this will be your main tool for manipulating different services that makes zensearch running.

Usage: 
- "start" to run zensearch services
- "stop"  stops all of the zensearch services
- "exit" to quit from zensearch or just ctrl-c 
- "build" for building services
- "node-install" installing node specific dependencies dependencies

\n`)

}
