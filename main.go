package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

var fileExec = [][]string{{"express-server", "node", "express-server/src/index.ts"}}
var errArr = [][]string{}
var runCmds = [][]string{
	{"express", "node", "./database/dist/index.js"},
	{"database", "node", "./express-server/dist/index.js"},
	{"crawler", "./crawler/crawler"},
	{"search-engine", "./search-engine/search-engine"},
}

var rabbitmqContConfig = DockerContainerConfig{
	HostPorts:      HostPorts{"5672", "15672"},
	ContainerPorts: ContainerPorts{{"5672", "5672"}, {"15672", "15672"}},
	Name:           "zensearch-cli-rabbitmq",
}
var dockerContainerConf = []DockerContainerConfig{rabbitmqContConfig}

func main() {
	scanner := bufio.NewScanner(os.Stdin)

loop:
	for {
		fmt.Printf("zensearch> ")
		scanner.Scan()
		text := scanner.Text()
		input := strings.Trim(text, " ")
		switch input {
		case "start":
			startServices(runCmds)
			break
		case "stop":
			// send kill signal to each process
			fmt.Printf("Input received %s:\n", input)
			fmt.Printf("Stopping services...\n")
			break loop
		case "build":
			fmt.Printf("zensearch: Building...\n")
			// runCommands(buildCmds, &errArr)
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
