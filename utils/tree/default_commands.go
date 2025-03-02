package tree

var npmNodes = []string{"npm", "install"}
var golangNodes = []string{"go", "build", "-C"}
var defaultCmdNodes = [][]string{npmNodes, golangNodes}

// init go and npm default build commands
func initDefaultBuildExec(r *RootNode) {
	for i := range defaultCmdNodes {
		r.children = append(r.children, &Node{value: defaultCmdNodes[i][0]})
		defaultNodesInit(r.children[i], defaultCmdNodes[i], 0)
	}
}

// appends the next node
// nc for node count
func defaultNodesInit(n *Node, defaultNodes []string, nc int) *Node {
	// stop defaultNodes has been exhausted
	nc++
	if nc >= len(defaultNodes) {
		return nil
	}
	n.next = &Node{value: defaultNodes[nc]}
	return defaultNodesInit(n.next, defaultNodes, nc)
}
