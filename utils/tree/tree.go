package tree

import (
	"fmt"
)

type DecisionTree struct {
	root     *RootNode
	cmdArr   []string
	execNode *Node
}

type RootNode struct {
	value    string
	children []*Node
}

type Node struct {
	value string
	next  *Node
}

func NewTree(initChildren func(*RootNode)) *DecisionTree {
	rootNode := &RootNode{children: []*Node{}}
	initChildren(rootNode)
	return &DecisionTree{rootNode, []string{}, nil}
}

// bruh polynomial time?
func (t DecisionTree) printNodes() {
	hLine := func(n int) string {
		tmp := ""
		for range n {
			tmp += "-"
		}
		return tmp
	}
	for _, child := range t.root.children {
		currentNode := child
		fmt.Printf("|-%s\n", currentNode.value)
		l := 2
		for currentNode.next != nil {
			currentNode = currentNode.next
			fmt.Printf("|%s %s\n", hLine(l), currentNode.value)
			l++
		}
		fmt.Println("")
	}
}

func (t *DecisionTree) buildCmd(arr []string) (string, error) {
	for _, child := range t.root.children {
		if child.value == t.peekNext(arr) {
			fmt.Printf("tree: got %s command\n", child.value)
			t.execNode = child
			return t.appendCommands(arr), nil
		}
	}
	return "", fmt.Errorf("tree: command exec does not exists (check the first element if that command is correctly installed to be able to run\n)")
}

func (t *DecisionTree) appendCommands(arr []string) string {
	tmp := arr[0] // service name

	cChild := t.execNode

	for cChild != nil {
		tmp += " " + cChild.value
		cChild = cChild.next
	}

	for _, args := range arr[2:] {
		tmp += " " + args
	}
	return tmp
}

func (t *DecisionTree) peekNext(arr []string) string {
	if len(arr) > 0 {
		return arr[1] // command name
	}
	return ""
}
