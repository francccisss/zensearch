package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
)

type StdError struct {
	src   string
	value string
}

func NewStdError(src string) StdError {
	return StdError{src: src}
}

func (se StdError) Error() string {
	return fmt.Sprintf("%s: ERROR %s", se.src, se.value)
}

func (se *StdError) addError(value string) {
	se.value = value
}

// TODO make sure docker services are running first
// use go routines and wait for state changes
// need to start containerized services first eg: selenium/rabbitmq
func startServices(pctx context.Context, commands [][]string) {
	fmt.Println("zensearch: Starting services...")
	ctx, cancelFunc := context.WithCancel(pctx)
	errChan := make(chan error)

	var wg sync.WaitGroup

	for _, contConfig := range dockerContainerConf {
		wg.Add(1)
		go runningDockerService(ctx, &wg, contConfig, errChan)
	}
	wg.Wait()
	go func() {
		for _, command := range commands {
			cmd := exec.Command(command[1], command[2:]...)
			go runningService(ctx, cmd, errChan, command[0])
		}
	}()

	// main context listener
	go func() {
		fmt.Println("zensearch: spawning context listener")
		select {
		case err := <-errChan:
			fmt.Println(err.Error())
			cancelFunc()
			fmt.Println("zensearch: cancelling all services...")
			fmt.Println("zensearch: services cancelled")
		case <-ctx.Done():
			// TODO CLEAN UP SERVICES HERE
			fmt.Println("zensearch: cleaning up services...")
			fmt.Println("zensearch: services stopped")
		}
	}()

}

func runningDockerService(ctx context.Context, wg *sync.WaitGroup, contConfig DockerContainerConfig, errChan chan error) {
	defer wg.Done()

	cont := NewContainer(contConfig.Name, contConfig.HostPorts, contConfig.ContainerPorts)

	go func() {
		<-ctx.Done()
		fmt.Printf("Docker: shutting down %s container...\n", contConfig.Name)
		cont.Stop(ctx)
		fmt.Printf("Docker: %s container stopped\n", contConfig.Name)
	}()

	err := cont.Run(ctx, "rabbitmq", "4.0-management")
	if err != nil {
		errChan <- err
		return
	}
}
func runningService(ctx context.Context, cmd *exec.Cmd, errChan chan error, cmdName string) {
	newStdErr := NewStdError(cmdName)
	stdout, err := cmd.StdoutPipe()
	stderr, err := cmd.StderrPipe()

	go func() {
		<-ctx.Done()
		fmt.Printf("%s shutting down process...\n", cmdName)
	}()

	if err != nil {
		fmt.Printf("zensearch: ERROR unable to set up stdout for process %s\n", cmdName)
		newStdErr.addError(err.Error())
		errChan <- newStdErr
		return
	}
	err = cmd.Start()
	if err != nil {
		fmt.Printf("zensearch: ERROR unable to start process %s\n", cmdName)
		newStdErr.addError(err.Error())
		errChan <- newStdErr
		return
	}
	// for handling stderr
	go func() {
		readErrors, err := io.ReadAll(stderr)
		if err != nil {
			fmt.Printf("zensearch: ERROR unable to read errors from process %s\n", cmdName)
		}
		newStdErr.addError(string(readErrors))
		errChan <- newStdErr
	}()

	io.Copy(os.Stdout, stdout)
	// use stderr to check the state of the process
	err = cmd.Wait()
	if err != nil {
		newStdErr.addError(err.Error())
		errChan <- newStdErr
	}
}

func runCommands(commands [][]string, errArr *[][]string) {

	fmt.Println("NOTICE FOR DATABASE SERVICE: make sure you have sqlite3 installed on your system!")
	for _, command := range commands {
		cmd := exec.Command(command[1], command[2:]...)
		stdErr, err := cmd.StderrPipe()
		stdOut, err := cmd.StdoutPipe()
		err = cmd.Start()
		io.Copy(os.Stdout, stdOut)
		if err != nil {
			fmt.Println("Error: cannot run command")
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
				fmt.Printf("%s> %s\n", command[0], string(readStdErr))
				panic(err)
			default:
				panic(err)
			}
			*errArr = append(*errArr, []string{command[0], err.Error()})
		}

		for _, str := range command {
			if str == "install" {
				fmt.Printf("%s: installing dependencies for %s service...\n", command[0], command[0])
			} else if str == "build" {
				fmt.Printf("%s: building %s service...\n", command[0], command[0])
			}
		}
		if err != nil {
			fmt.Println(err.Error())
			*errArr = append(*errArr, []string{command[0], err.Error()})
		}
		cmd.Wait()

		for _, str := range command {
			if str == "install" {
				fmt.Printf("%s: dependencies successfully installed\n", command[0])
			} else if str == "build" {
				fmt.Printf("%s: build successful\n", command[0])
			}
		}
	}

}

func help() {
	fmt.Printf(`
Welcome to zensearch cli this will be your main tool for manipulating different services that makes zensearch running.

Usage: 
- "start" to build and run zensearch
- "stop"  stops all of the zensearch services
- "build" for building and installing dependencies

For database handling, for now you can use the system installed sqlite3 for manipulating your database located in the '/database/website_collection.db' if you know how to use sqlite3 then you know what to do, but for others please read the sqlite3 docs :D

`)
	fmt.Println("")

}
