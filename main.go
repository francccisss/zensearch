package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

var fileExec = [][]string{{"express-server", "node", "express-server/src/index.ts"}}
var buildCmd = [][]string{
	{"express-server", "npm", "install", "./express-server"},
	{"database", "npm", "install", "./database"},
	{"crawler", "go", "build", "-C", "./crawler/"},
	{"search-engine", "go", "build", "-C", "./search-engine/"},
}
var errArr = [][]string{}

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
			fmt.Printf("Input received %s:\n", input)
			docker := "docker run -it --rm --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:4.0-management"
			cmd := exec.Command(docker)
			stdOut, err := cmd.StdoutPipe()
			cmd.Start()

			if err != nil {
				fmt.Println("Error: cannot run command")
				switch e := err.(type) {
				case *exec.Error:
					fmt.Println("failed executing:", err)
					break
				case *exec.ExitError:
					fmt.Println(err.Error())
					fmt.Println("command exit rc =", e.ExitCode())
					panic(err)
				default:
					panic(err)
				}
			}

			readOut, err := io.ReadAll(stdOut)
			if err != nil {
				fmt.Println("Unable to read stdout")
				fmt.Println(err.Error())
				errArr = append(errArr, []string{"docker", err.Error()})
			}
			fmt.Println(string(readOut))
			cmd.Wait()

			break
		case "stop":
			// send kill signal to each process
			fmt.Printf("Input received %s:\n", input)
			fmt.Printf("Stopping services...\n")
			break loop
		case "build":
			fmt.Printf("zensearch: Building...\n")
			build(buildCmd, &errArr)
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
