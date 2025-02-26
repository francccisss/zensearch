package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	// "time"
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
			break
		case "stop":
			fmt.Printf("Input received %s:\n", input)
			fmt.Printf("Stopping process...\n")
			break loop
		case "build":
			fmt.Printf("Input received %s:\n", input)
			fmt.Printf("Building...\n")

			for _, file := range fileExec {

				cmd := exec.Command(file[1], file[2])
				err := cmd.Run()
				if err != nil {
					fmt.Println("Error: cannot run command")
					errArr = append(errArr, []string{file[0], err.Error()})
					continue
				}
				go func(c *exec.Cmd) {
					b := []byte{}
					_, err := c.Stdout.Write(b)
					_, err = c.Stderr.Write(b)
					if err != nil {
						errArr = append(errArr, []string{file[0], err.Error()})
					}
					fmt.Printf("%s > %s\n", file[0], b)
				}(cmd)

				// time.Sleep(10 * time.Second)
			}
			printErrors(&errArr)
			break
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
