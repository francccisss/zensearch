package tree

import (
	"fmt"
	"testing"
)

var buildCmd = [][]string{
	{"express-service", "npm", "./database"},
	{"database-service", "npm", "./express-server"},
	{"crawler", "go", "./crawler/"},
	{"search-engine", "go", "./search-engine/"},
}

func TestBuildCmdTrail(t *testing.T) {
	fmt.Println("Building cmd trail")
	dt := NewTree(initDefaultBuildExec)
	if len(dt.root.children) == 0 {
		fmt.Println("Default nodes not added")
		t.FailNow()
	}

	dt.printNodes()
	for _, cmds := range buildCmd {

		cmd, err := dt.buildCmd(cmds)
		if err != nil {
			fmt.Println(err.Error())
			t.FailNow()
		}
		fmt.Printf("tree: cmd %s\n", cmd)
	}
	fmt.Print("tree: test done")
}

func TestBuildDefaultCmdTree(t *testing.T) {
	fmt.Println("Testing default cmd tree")
	dt := NewTree(initDefaultBuildExec)
	if len(dt.root.children) == 0 {
		fmt.Println("Default nodes not added")
		t.FailNow()
	}
	dt.printNodes()
}

func TestDefaultNodeCommandsInit(t *testing.T) {
	fmt.Println("Testing default node initilization")
	var npmNodes = []string{"npm", "install"}
	var golangNodes = []string{"go", "build", "-C"}
	tTables := [][]string{npmNodes, golangNodes}

	for i := range tTables {
		startingNode := &Node{value: tTables[i][0]}
		defaultNodesInit(startingNode, tTables[i], 0)
		if startingNode.next == nil {
			fmt.Println("lower node not appended to starting node")
			t.FailNow()
		}
		fmt.Printf("%+v\n", startingNode)
		fmt.Printf("|- %+v\n", startingNode.next)
		fmt.Printf("|-- %+v\n\n", startingNode.next.next)
	}
}
