package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/docker/go-connections/nat"
)

var runCmds = [][]string{
	{"express", "node", "./express-server/dist/index.js"},
	{"database", "node", "./database/dist/index.js"},
	// {"crawler", "./crawler/crawler-bin"},
	// {"search-engine", "./search-engine/search-engine-bin"},
}

var buildCmds = [][]string{
	{"express", "npm", "run", "build", "--prefix", "./express-server"},
	{"database", "npm", "run", "build", "--prefix", "./database"},
	{"crawler", "go", "build", "-C", "./crawler/", "-o", "crawler-bin"},
	{"search-engine", "go", "build", "-C", "./search-engine/", "-o", "search-engine-bin"},
}

var npmInstall = [][]string{
	{"express", "npm", "install", "express-server/"},
	{"database", "npm", "install", "database/"},
}

// NOTICE: Make sure every image is compatible with the host system, arm64, linux, windows

var RabbitmqConfig = DockerContainerConfig{
	HostPorts:      HostPorts{"5672/tcp", "15672/tcp"},
	ContainerPorts: nat.PortSet{"5672/tcp": struct{}{}, "15672/tcp": struct{}{}},
	Name:           "zensearch-cli-rabbitmq",
	ImageName:      "rabbitmq",
	Tag:            "4.0-management",
}

// TODO use options method for optional arguments still dont know how to do that
var SeleniumConfig = DockerContainerConfig{
	ImageName:      "selenium/standalone-chromium",
	Tag:            "latest",
	HostPorts:      HostPorts{"4444/tcp", "7900/tcp"},
	ContainerPorts: nat.PortSet{"4444/tcp": struct{}{}, "7900/tcp": struct{}{}},
	Name:           "zensearch-cli-selenium",
	ShmSize:        2 * 1024 * 1024 * 1024,
	Env:            []string{"SE_NODE_MAX_SESSIONS=5"},
}
var dockerContainerConf = []DockerContainerConfig{RabbitmqConfig, SeleniumConfig}

func main() {
	// for reassigning cancel func if services has started
	var contextCancel context.CancelFunc
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go CLILoop(contextCancel)

	<-sigChan
	contextCancel()

	time.Sleep(time.Second * 3)
	fmt.Println("Exit")
}

func printErrors(errArr *[][]string) {
	for _, err := range *errArr {
		fmt.Printf("%s: ERROR %s\n", err[0], err[1])
	}
	*errArr = nil
}

func CLILoop(contextCancel context.CancelFunc) {

	scanner := bufio.NewScanner(os.Stdin)
loop:
	for {
		fmt.Printf("zensearch > ")
		scanner.Scan()
		text := scanner.Text()
		input := strings.Trim(text, " ")
		switch input {
		case "start":
			ctx, cancel := context.WithCancel(context.Background())
			contextCancel = cancel
			startServices(ctx, runCmds)
		case "stop":
			// send kill signal to each process

			fmt.Println("Stopping zensearch")
			fmt.Println("Stopping services...")
			if contextCancel != nil {
				contextCancel()
			}
			fmt.Println("Zensearch stopped")
		case "exit":
			// send kill signal to each process
			fmt.Println("exiting zensearch")
			fmt.Println("Stopping services...")
			if contextCancel != nil {
				contextCancel()
			}
			fmt.Println("Exit")
			break loop
		case "build":
			fmt.Printf("zensearch: Building...\n")
			runCommands(buildCmds)
		case "node-install":
			fmt.Printf("zensearch: installing node dependencies...\n")
			runCommands(npmInstall)
		default:
			help()
		}
	}
}
