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
)

var errArr = [][]string{}
var runCmds = [][]string{
	{"express", "node", "./express-server/dist/index.js"},
	{"database", "node", "./database/dist/index.js"},
	{"crawler", "./crawler/crawler-bin"},
	{"search-engine", "./search-engine/search-engine-bin"},
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

var rabbitmqContConfig = DockerContainerConfig{
	HostPorts:      HostPorts{"5672", "15672"},
	ContainerPorts: ContainerPorts{{"5672", "5672"}, {"15672", "15672"}},
	Name:           "zensearch-cli-rabbitmq",
	ImageName:      "rabbitmq",
	Tag:            "4.0-management",
}

// TODO use options method for optional arguments still dont know how to do that
var seleniumContConfig = DockerContainerConfig{
	ImageName:      "selenium/standalone-chromium",
	Tag:            "latest",
	HostPorts:      HostPorts{"4444", "7900"},
	ContainerPorts: ContainerPorts{{"4444", "4444"}, {"7900", "7900"}},
	Name:           "zensearch-cli-selenium-multi-arch",
	ShmSize:        2 * 1024 * 1024 * 1024,
	Env:            []string{"SE_NODE_MAX_SESSIONS=5"},
}
var dockerContainerConf = []DockerContainerConfig{rabbitmqContConfig, seleniumContConfig}

func main() {
	// for reassigning cancel func if services has started
	var contextCancel context.CancelFunc
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go CLILoop(contextCancel)
	printErrors(&errArr)

	// TODO: this does not look nice at all
	<-sigChan
	if contextCancel != nil {
		contextCancel()
	}
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
		fmt.Printf("zensearch> ")
		scanner.Scan()
		text := scanner.Text()
		input := strings.Trim(text, " ")
		switch input {
		case "start":
			ctx, cancel := context.WithCancel(context.Background())
			contextCancel = cancel
			startServices(ctx, runCmds)
			break
		case "stop":
			// send kill signal to each process

			fmt.Println("Stopping zensearch")
			fmt.Println("Stopping services...")
			if contextCancel != nil {
				contextCancel()
			}
			fmt.Println("Zensearch stopped")
			break
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
			runCommands(buildCmds, &errArr)
			break
		case "node-install":
			fmt.Printf("zensearch: installing node dependencies...\n")
			runCommands(npmInstall, &errArr)
			break
		case "help":
			help()
		default:
			break
		}
	}
}
