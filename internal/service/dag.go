package service

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
)

// Node represents a DAG node with parent reference and content
type Node struct {
	ParentID string
	Content  string
}

// Hash generates a deterministic hash for the node
func (n *Node) Hash() string {
	data := n.ParentID + n.Content
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// TagDef defines a tag type with metadata
type TagDef struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}

// DAG represents a directed acyclic graph
type DAG struct {
	Nodes   map[string]*Node    // hash -> node
	Root    *Node               // reference to root node
	TagDefs map[string]*TagDef  // tag ID -> tag definition
	Tags    map[string][]string // tag ID -> list of node hashes
}

// New creates a new DAG instance
func New() *DAG {
	return &DAG{
		Nodes:   make(map[string]*Node),
		TagDefs: make(map[string]*TagDef),
		Tags:    make(map[string][]string),
	}
}

// AddNode adds a node to the DAG and returns its hash
func (d *DAG) AddNode(node *Node) string {
	hash := node.Hash()
	d.Nodes[hash] = node

	// If this is the first node (no parent), set it as root
	if node.ParentID == "" && d.Root == nil {
		d.Root = node
	}

	return hash
}

// Context recursively walks back to root and concatenates content from parent to child
func (d *DAG) Context(id string) (string, error) {
	node, exists := d.Nodes[id]
	if !exists {
		return "", errors.New("node not found")
	}

	// Build path from current node to root
	var path []*Node
	current := node

	for {
		path = append([]*Node{current}, path...) // prepend to maintain order

		if current.ParentID == "" {
			// Reached root
			break
		}

		parent, exists := d.Nodes[current.ParentID]
		if !exists {
			return "", fmt.Errorf("parent node %s not found", current.ParentID)
		}

		current = parent
	}

	// Concatenate content from root to current node
	var context string
	for _, n := range path {
		context += n.Content
	}

	return context, nil
}

// AddTag adds or updates a tag definition
func (d *DAG) AddTag(tagDef *TagDef) {
	d.TagDefs[tagDef.ID] = tagDef
	if _, exists := d.Tags[tagDef.ID]; !exists {
		d.Tags[tagDef.ID] = []string{}
	}
}

// GetTag returns a tag definition
func (d *DAG) GetTag(tagID string) (*TagDef, error) {
	tagDef, exists := d.TagDefs[tagID]
	if !exists {
		return nil, errors.New("tag not found")
	}
	return tagDef, nil
}

// GetAllTags returns all tag definitions
func (d *DAG) GetAllTags() []*TagDef {
	tags := make([]*TagDef, 0, len(d.TagDefs))
	for _, tag := range d.TagDefs {
		tags = append(tags, tag)
	}
	return tags
}

// TagNode adds a tag to a node
func (d *DAG) TagNode(nodeID, tagID, description string) error {
	if _, exists := d.Nodes[nodeID]; !exists {
		return errors.New("node not found")
	}

	// Create or update tag definition
	d.AddTag(&TagDef{
		ID:          tagID,
		Description: description,
	})

	// Check if node already has this tag
	nodeList := d.Tags[tagID]
	for _, nid := range nodeList {
		if nid == nodeID {
			// Already tagged, just return
			return nil
		}
	}

	// Add node to tag list
	d.Tags[tagID] = append(nodeList, nodeID)
	return nil
}

// UntagNode removes a tag from a node
func (d *DAG) UntagNode(nodeID, tagID string) error {
	if _, exists := d.Nodes[nodeID]; !exists {
		return errors.New("node not found")
	}

	nodeList, exists := d.Tags[tagID]
	if !exists {
		return errors.New("tag not found")
	}

	// Remove node from tag list
	for i, nid := range nodeList {
		if nid == nodeID {
			d.Tags[tagID] = append(nodeList[:i], nodeList[i+1:]...)
			return nil
		}
	}

	return errors.New("node does not have this tag")
}

// GetNodesWithTag returns all node IDs with a specific tag
func (d *DAG) GetNodesWithTag(tagID string) ([]string, error) {
	nodeList, exists := d.Tags[tagID]
	if !exists {
		return nil, errors.New("tag not found")
	}
	return nodeList, nil
}

// ProcessPrompt is a stub that simulates an LLM call
func (d *DAG) ProcessPrompt(contextStr, prompt string) string {
	// Stub implementation - just return a formatted response
	return fmt.Sprintf("Response to: %s (context length: %d)", prompt, len(contextStr))
}
