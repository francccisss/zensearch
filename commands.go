package main

import (
	"fmt"
	"io"
	"os/exec"
)

func build(commands [][]string, errArr *[][]string) {

	for _, command := range commands {
		cmd := exec.Command(command[1], command[2:]...)
		stdErr, err := cmd.StderrPipe()
		// stdOut, err := cmd.StdoutPipe()
		err = cmd.Start()
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
		fmt.Printf("%s: building %s service...\n", command[0], command[0])
		if command[0] == "database" {
			fmt.Println("NOTICE FOR DATABASE SERVICE: make sure you have sqlite3 installed on your system!")
		}
		// readStdOut, err := io.ReadAll(stdOut)
		if err != nil {
			fmt.Println(err.Error())
			*errArr = append(*errArr, []string{command[0], err.Error()})
		}
		cmd.Wait()
		switch command[1] {
		case "go":
			fmt.Printf("%s: build successful\n", command[0])
			break
		case "npm":
			fmt.Printf("%s: installed node dependencies\n", command[0])
			break
		}
	}

}
