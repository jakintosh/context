package service

import (
	"testing"
)

func TestNodeHash(t *testing.T) {
	node1 := &Node{
		ParentID: "parent123",
		Content:  "test content",
	}

	node2 := &Node{
		ParentID: "parent123",
		Content:  "test content",
	}

	node3 := &Node{
		ParentID: "parent123",
		Content:  "different content",
	}

	hash1 := node1.Hash()
	hash2 := node2.Hash()
	hash3 := node3.Hash()

	if hash1 != hash2 {
		t.Errorf("identical nodes should have same hash: %s != %s", hash1, hash2)
	}

	if hash1 == hash3 {
		t.Errorf("different nodes should have different hashes")
	}

	if len(hash1) != 64 {
		t.Errorf("expected SHA256 hash length of 64, got %d", len(hash1))
	}
}

func TestAddNode(t *testing.T) {
	dag := New()

	// Add root node
	root := &Node{
		ParentID: "",
		Content:  "root content",
	}
	rootHash := dag.AddNode(root)

	if dag.Root != root {
		t.Error("root node not set correctly")
	}

	if _, exists := dag.Nodes[rootHash]; !exists {
		t.Error("node not added to nodes map")
	}

	// Add child node
	child := &Node{
		ParentID: rootHash,
		Content:  "child content",
	}
	childHash := dag.AddNode(child)

	if _, exists := dag.Nodes[childHash]; !exists {
		t.Error("child node not added to nodes map")
	}

	if rootHash == childHash {
		t.Error("parent and child should have different hashes")
	}
}

func TestContext(t *testing.T) {
	dag := New()

	// Create a chain: root -> child1 -> child2
	root := &Node{
		ParentID: "",
		Content:  "A",
	}
	rootHash := dag.AddNode(root)

	child1 := &Node{
		ParentID: rootHash,
		Content:  "B",
	}
	child1Hash := dag.AddNode(child1)

	child2 := &Node{
		ParentID: child1Hash,
		Content:  "C",
	}
	child2Hash := dag.AddNode(child2)

	// Test context at root
	context, err := dag.Context(rootHash)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if context != "A" {
		t.Errorf("expected 'A', got '%s'", context)
	}

	// Test context at child1
	context, err = dag.Context(child1Hash)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if context != "AB" {
		t.Errorf("expected 'AB', got '%s'", context)
	}

	// Test context at child2
	context, err = dag.Context(child2Hash)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if context != "ABC" {
		t.Errorf("expected 'ABC', got '%s'", context)
	}

	// Test non-existent node
	_, err = dag.Context("nonexistent")
	if err == nil {
		t.Error("expected error for non-existent node")
	}
}

func TestTagManagement(t *testing.T) {
	dag := New()

	// Add a node
	node := &Node{
		ParentID: "",
		Content:  "test",
	}
	nodeHash := dag.AddNode(node)

	// Test AddTag
	tagDef := &TagDef{
		ID:          "important",
		Description: "Important items",
	}
	dag.AddTag(tagDef)

	// Verify tag was added
	retrievedTag, err := dag.GetTag("important")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if retrievedTag.ID != "important" || retrievedTag.Description != "Important items" {
		t.Errorf("tag not added correctly")
	}

	// Test GetAllTags
	allTags := dag.GetAllTags()
	if len(allTags) != 1 {
		t.Errorf("expected 1 tag, got %d", len(allTags))
	}

	// Test TagNode
	err = dag.TagNode(nodeHash, "important", "Important items")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify node was tagged
	nodes, err := dag.GetNodesWithTag("important")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(nodes) != 1 || nodes[0] != nodeHash {
		t.Errorf("node not tagged correctly")
	}

	// Test tagging with non-existent node
	err = dag.TagNode("nonexistent", "important", "Important items")
	if err == nil {
		t.Error("expected error for non-existent node")
	}

	// Test UntagNode
	err = dag.UntagNode(nodeHash, "important")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	nodes, err = dag.GetNodesWithTag("important")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(nodes) != 0 {
		t.Errorf("expected 0 nodes after untagging, got %d", len(nodes))
	}

	// Test untagging non-existent tag
	err = dag.UntagNode(nodeHash, "nonexistent")
	if err == nil {
		t.Error("expected error for non-existent tag")
	}
}

func TestTagNodeDuplicatePrevention(t *testing.T) {
	dag := New()

	node := &Node{
		ParentID: "",
		Content:  "test",
	}
	nodeHash := dag.AddNode(node)

	// Tag the node twice
	dag.TagNode(nodeHash, "test", "Test tag")
	dag.TagNode(nodeHash, "test", "Test tag")

	nodes, err := dag.GetNodesWithTag("test")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(nodes) != 1 {
		t.Errorf("expected 1 node after duplicate tagging, got %d", len(nodes))
	}
}

func TestProcessPrompt(t *testing.T) {
	dag := New()

	response := dag.ProcessPrompt("some context", "test prompt")

	if response == "" {
		t.Error("expected non-empty response from ProcessPrompt")
	}

	if len(response) == 0 {
		t.Error("ProcessPrompt should return a non-empty string")
	}
}
