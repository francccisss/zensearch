package main

import (
	"testing"
)

func TestNpmInstallCmd(t *testing.T) {
	runCommands(npmInstall, &errArr)
}
func TestBuildCmd(t *testing.T) {
	runCommands(buildCmds, &errArr)
}
func TestRunCmd(t *testing.T) {
	runCommands(runCmds, &errArr)
}
