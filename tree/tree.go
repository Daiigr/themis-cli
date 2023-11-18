package tree

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"themis-cli/parser"
)

const (
	baseURL = "https://themis.housing.rug.nl"
)

// AssignmentNode represents a node in a tree structure used for storing assignments.
type AssignmentNode struct {
	Parent   *AssignmentNode
	Name     string
	URL      string
	children []*AssignmentNode
}

// AppendChild appends a child node to the parent node.
// It sets the parent of the child node and adds the child node to the parent's list of children.
func (n *AssignmentNode) AppendChild(c *AssignmentNode) {
	log.Default().Println(fmt.Sprintf("Appending child %s to parent %s", c.Name, n.Name))
	c.Parent = n
	n.children = append(n.children, c)
}

// BuildAssignmentNode creates a new AssignmentNode with the specified parent, name, and URL.
// It logs a message indicating the node being built and returns the created node.
func BuildAssignmentNode(parent *AssignmentNode, name string, url string) *AssignmentNode {
	log.Default().Println(fmt.Sprintf("Building node %s", name))
	node := &AssignmentNode{
		Name:   name,
		URL:    url,
		Parent: parent,
	}
	return node
}

// BuildRootAssignmentNode creates a root assignment node with the given name and URL.
// It calls the BuildAssignmentNode function with a nil parent node.
func BuildRootAssignmentNode(name string, url string) *AssignmentNode {
	return BuildAssignmentNode(nil, name, url)
}

func PullAssignmentsFromThemisAndBuildTree(client *http.Client, URL string, rootNode *AssignmentNode, depth int) (*AssignmentNode, error) {

	// get assignments on page
	assignments, err := parser.GetAssignmentsOnPage(client, URL)
	if err != nil {
		return nil, fmt.Errorf("error getting assignments on page: %v", err)
	}

	// build assignment nodes
	for _, assignment := range assignments {
		assignmentNode := BuildAssignmentNode(rootNode, assignment["name"], assignment["url"])
		rootNode.AppendChild(assignmentNode)
	}

	// build tree
	if depth > 0 {
		for _, child := range rootNode.children {
			child, err = PullAssignmentsFromThemisAndBuildTree(client, child.URL, child, depth-1)
			if err != nil {
				return nil, fmt.Errorf("error building tree: %v", err)
			}
		}
	}

	return rootNode, nil
}

func SaveAssignmentTreeToJSON(rootNode *AssignmentNode, depth int) error {
	file, err := os.Create("assignment_tree.json")
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}

	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			log.Printf("error closing file: %v", closeErr)
		}
	}()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ") // format output with 4 spaces

	err = encodeAssignmentTree(encoder, rootNode, depth)
	if err != nil {
		return fmt.Errorf("error encoding assignment tree: %v", err)
	}

	return nil
}

func encodeAssignmentTree(encoder *json.Encoder, node *AssignmentNode, depth int) error {
	if depth < 0 {
		return nil
	}

	err := encoder.Encode(node)
	if err != nil {
		return fmt.Errorf("error encoding node %s: %v", node.Name, err)
	}

	for _, child := range node.children {
		err = encodeAssignmentTree(encoder, child, depth-1)
		if err != nil {
			return err
		}
	}

	return nil
}

func BuildTreeFromJSON(fileName string) (rootNode *AssignmentNode) {

}
