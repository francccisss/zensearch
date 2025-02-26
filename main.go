package main

import (
	"bufio"
	"fmt"
	"os"
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
			break
		case "stop":
			fmt.Printf("Input received %s:\n", input)
			fmt.Printf("Stopping process...\n")
			break loop
		case "build":
			fmt.Printf("zensearch: Building...\n")
			build(buildCmd, &errArr)
			break
		case "help":
			fmt.Printf(`
Welcome to zensearch cli this will be your main tool for manipulating different services that makes zensearch running.

Usage: 
- "start" to build and run zensearch
- "stop"  stops all of the zensearch services
- "build" for building and installing dependencies

For database handling, for now you can use the system installed sqlite3 for manipulating your database located in the '/database/website_collection.db' if you know how to use sqlite3 then you know what to do, but for others please read the sqlite3 docs :D

`)
			fmt.Println("")
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
