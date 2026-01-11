package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"git.sr.ht/~jakintosh/context/internal/service"
)

// Handler holds the DAG service and provides HTTP handlers
type Handler struct {
	dag *service.DAG
}

// NewHandler creates a new API handler
func NewHandler(dag *service.DAG) *Handler {
	return &Handler{dag: dag}
}

// GetTags handles GET /tags
func (h *Handler) GetTags(w http.ResponseWriter, r *http.Request) {
	tags := h.dag.GetAllTags()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tags)
}

// GetTag handles GET /tags/{id}
func (h *Handler) GetTag(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 2 {
		http.Error(w, "tag ID required", http.StatusBadRequest)
		return
	}
	tagID := parts[1]

	tag, err := h.dag.GetTag(tagID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tag)
}

// GetContext handles GET /context/{id}
func (h *Handler) GetContext(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 2 {
		http.Error(w, "node ID required", http.StatusBadRequest)
		return
	}
	nodeID := parts[1]

	context, err := h.dag.Context(nodeID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	response := map[string]string{"context": context}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// PostContext handles POST /context/{id}
func (h *Handler) PostContext(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 2 {
		http.Error(w, "node ID required", http.StatusBadRequest)
		return
	}
	parentID := parts[1]

	// Parse request body
	var req struct {
		Prompt string `json:"prompt"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Get context for the parent node
	contextStr, err := h.dag.Context(parentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Create node with user prompt
	promptNode := &service.Node{
		ParentID: parentID,
		Content:  req.Prompt,
	}
	promptNodeID := h.dag.AddNode(promptNode)

	// Call LLM stub with context
	responseText := h.dag.ProcessPrompt(contextStr, req.Prompt)

	// Create node with LLM response
	responseNode := &service.Node{
		ParentID: promptNodeID,
		Content:  responseText,
	}
	responseNodeID := h.dag.AddNode(responseNode)

	// Return response
	response := map[string]string{
		"id":       responseNodeID,
		"response": responseText,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// PutTag handles PUT /context/{id}/tag
func (h *Handler) PutTag(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 2 {
		http.Error(w, "node ID required", http.StatusBadRequest)
		return
	}
	nodeID := parts[1]

	// Parse request body
	var tag service.TagDef
	if err := json.NewDecoder(r.Body).Decode(&tag); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if tag.ID == "" || tag.Description == "" {
		http.Error(w, "id and description are required", http.StatusBadRequest)
		return
	}

	// Add tag to node
	if err := h.dag.TagNode(nodeID, tag.ID, tag.Description); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DeleteTag handles DELETE /context/{id}/tag
func (h *Handler) DeleteTag(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 2 {
		http.Error(w, "node ID required", http.StatusBadRequest)
		return
	}
	nodeID := parts[1]

	// Parse request body
	var req struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.ID == "" {
		http.Error(w, "tag id is required", http.StatusBadRequest)
		return
	}

	// Remove tag from node
	if err := h.dag.UntagNode(nodeID, req.ID); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}
