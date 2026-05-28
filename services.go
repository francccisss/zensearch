package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
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

// TODO: make it so that if users can use either docker desktop or just dockerd
// so that they configure the location of the host to their dockerd socket path
func startServices(pctx context.Context, commandsList [][]string) {
	newerr := NewError("Service Initialization")
	var mux sync.Mutex
	dockerMan, err := NewDockerManager(&mux)
	if err != nil {
		newerr.addError(err.Error())
		panic(newerr.Error())
	}
	fmt.Println("[ZENSEARCH]: Starting services...")

	var wg sync.WaitGroup

	fmt.Println("[DOCKER]: Starting up Docker Services")
	for _, contConfig := range dockerContainerConf {
		wg.Go(func() {
			runningDockerService(pctx, &dockerMan, contConfig)
		})
	}

	wg.Wait()

	fmt.Println("[ZENSEARCH]: Starting up Express & DB Services")
	for _, commands := range commandsList {
		wg.Go(func() {
			runningService(pctx, commands)
		})
	}
	fmt.Println("[ZENSEARCH]: Services started")
}

func runningDockerService(ctx context.Context, dockerMan *DockerManager, contConfig DockerContainerConfig) {
	newErr := NewError(string(contConfig.Name))

	dockerMan.NewDockerContainer(contConfig)

	err := dockerMan.PullImage(ctx, contConfig.ImageName+":"+contConfig.Tag)
	if err != nil {
		newErr.addError(err.Error())
		panic(newErr.Error())
	}
	// cancellation for specific service
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*120)
	defer cancel()
	go func() {
		<-ctx.Done()
		fmt.Printf("[DOCKER]: shutting down %s container...\n", contConfig.Name)
		if err := dockerMan.Stop(context.Background(), contConfig.Name); err != nil {
			newErr.addError(err.Error())
			fmt.Println(newErr.Error())
			return
		}
		fmt.Printf("[DOCKER]: %s container stopped\n", contConfig.Name)
	}()

	select {
	case err := <-dockerMan.Run(ctx, contConfig.Name, contConfig.ImageName+":"+contConfig.Tag):
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
func runningService(ctx context.Context, commands []string) {
	cmdName := commands[0]
	logName := strings.ToUpper(cmdName)

	newStdErr := NewError(cmdName)
	cmd := exec.Command(commands[1], commands[2:]...)
	// update the current working directory of each services
	// to run within its own directory to enable relative imports
	// for config files to be used.
	wdir, err := os.Getwd()
	if err != nil {
		fmt.Printf("[ZENSEARCH]: ERROR unable to get working directory %s\n", cmdName)
		newStdErr.addError(err.Error())
		panic(newStdErr.Error())
	}
	cmd.Dir = fmt.Sprintf("%s/%s/", wdir, cmdName)
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

		// bytes piped from services to zensearch stderr
		// needs to be read as they come in
		readerErr := bufio.NewReader(stderr)
		buf := make([]byte, 4096)
		for {
			n, err := readerErr.Read(buf)
			if err != nil {
				if err.Error() == io.EOF.Error() {
					return
				}
				fmt.Printf("[ZENSEARCH]: ERROR - %s - CAUSE %s\n", cmdName, err)
				return
			}
			// dont need to handle EOF since cmd.Wait already waits and cleans up resources and then exits the function if
			// io.EOF is encountered which can be a cause from an exit process
			if n > 0 {
				fmt.Printf("[%s]: LOG ERR - %s\n", logName, buf[:n])
			}
		}

	}()

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			fmt.Printf("[%s]: LOG - '%s'\n", logName, scanner.Text())
		}
	}()

	go func() {
		<-ctx.Done()
		fmt.Printf("[%s]: Shutting down\n", logName)
		if cmd.Process != nil {
			_ = cmd.Process.Signal(syscall.SIGTERM)
			fmt.Printf("[%s]: Closed, captured interrupt SIGTERM\n", logName)
		}
	}()

	err = cmd.Wait()
	switch e := err.(type) {
	case *exec.ExitError:
		status := e.ProcessState.Sys().(syscall.WaitStatus)
		if status.Signaled() {
			fmt.Printf("[%s]: Process received '%s' signal\n", logName, status.Signal())
			return
		}
		newStdErr.addError(err.Error())
		panic(newStdErr.Error())
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
