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

var seleniumContConfig = DockerContainerConfig{
	HostPorts:      HostPorts{"4444", "7900"},
	ContainerPorts: ContainerPorts{{"4444", "4444"}, {"7900", "7900"}},
	Name:           "zensearch-cli-selenium",
}
var dockerContainerConf = []DockerContainerConfig{rabbitmqContConfig, seleniumContConfig}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	// for reassigning cancel func if input is start
	var cancelFunc context.CancelFunc
	// TODO fix input field needs to be appended after every stdout
loop:
	for {
		fmt.Printf("zensearch> ")
		scanner.Scan()
		text := scanner.Text()
		input := strings.Trim(text, " ")
		switch input {
		case "start":
			ctx, cancel := context.WithCancel(context.Background())
			cancelFunc = cancel
			startServices(ctx, runCmds)
			fmt.Println("zensearch: services started")
			break
		case "stop":
			// send kill signal to each process
			fmt.Printf("Stopping services...\n")
			if cancelFunc != nil {
				cancelFunc()
			}
			break
		case "exit":
			// send kill signal to each process
			fmt.Printf("Input received %s:\n", input)
			fmt.Printf("Stopping services...\n")
			if cancelFunc != nil {
				cancelFunc()
			}
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
