package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"testing"
)

func TestNpmCmd(t *testing.T) {
	runCommands(npmInstall, &errArr)
}
func TestBuildCmd(t *testing.T) {
	runCommands(buildCmds, &errArr)
}

func TestRunCmd(t *testing.T) {
	startServices(runCmds)
}

func TestSignalINT(t *testing.T) {
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT)
	done := make(chan bool, 1)

	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		done <- true
	}()

	fmt.Println("awaiting signal")
	<-done
	fmt.Println("exiting")
}
