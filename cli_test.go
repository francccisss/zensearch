package main

import (
	"testing"
)

var buildCmds = [][]string{
	{"express", "npm", "run", "build", "--prefix", "./database"},
	{"database", "npm", "run", "build", "--prefix", "./express-server"},
	{"crawler", "go", "build", "-C", "./crawler/"},
	{"search-engine", "go", "build", "-C", "./search-engine/"},
}

var npmInstall = [][]string{
	{"express", "npm", "install", "database/"},
	{"database", "npm", "install", "express-server/"},
}

func TestNpmCmd(t *testing.T) {
	runCommands(npmInstall, &errArr)
}
func TestBuildCmd(t *testing.T) {
	runCommands(buildCmds, &errArr)
}

func TestRunCmd(t *testing.T) {
	startServices(runCmds)
}
