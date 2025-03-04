package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
)

var errArr = [][]string{}
var runCmds = [][]string{
	{"express", "node", "./express-server/dist/index.js"},
	{"database", "node", "./database/dist/index.js"},
	// {"crawler", "./crawler/crawler"},
	// {"search-engine", "./search-engine/search-engine"},
}

var buildCmds = [][]string{
	{"express", "npm", "run", "build", "--prefix", "./express-server"},
	{"database", "npm", "run", "build", "--prefix", "./database"},
	// {"crawler", "go", "build", "-C", "./crawler/"},
	// {"search-engine", "go", "build", "-C", "./search-engine/"},
}

var npmInstall = [][]string{
	{"express", "npm", "install", "express-server/"},
	{"database", "npm", "install", "database/"},
}

var rabbitmqContConfig = DockerContainerConfig{
	HostPorts:      HostPorts{"5672", "15672"},
	ContainerPorts: ContainerPorts{{"5672", "5672"}, {"15672", "15672"}},
	Name:           "zensearch-cli-rabbitmq",
}
var dockerContainerConf = []DockerContainerConfig{rabbitmqContConfig}

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	ctx, cancelFunc := context.WithCancel(context.Background())
loop:
	for {
		fmt.Printf("zensearch> ")
		scanner.Scan()
		text := scanner.Text()
		input := strings.Trim(text, " ")
		switch input {
		case "start":
			startServices(ctx, runCmds)
			fmt.Println("zensearch: services started")
			break
		case "stop":
			// send kill signal to each process
			fmt.Printf("Stopping services...\n")
			cancelFunc()
			break
		case "exit":
			// send kill signal to each process
			fmt.Printf("Input received %s:\n", input)
			fmt.Printf("Stopping services...\n")
			cancelFunc()
			break loop
		case "build":
			fmt.Printf("zensearch: Building...\n")
			runCommands(buildCmds, &errArr)
			break
		case "node-install":
			fmt.Printf("zensearch: Building...\n")
			runCommands(buildCmds, &errArr)
			break
		case "help":
			help()
		default:
			break
		}
	}
	printErrors(&errArr)
	fmt.Println("Process stopped")
}

func printErrors(errArr *[][]string) {
	for _, err := range *errArr {
		fmt.Printf("%s: ERROR %s\n", err[0], err[1])
	}
	*errArr = nil
}
func addError(name string, err error) {
}
