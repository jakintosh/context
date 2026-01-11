package main

import (
	"log"
	"net/http"
	"strings"

	"git.sr.ht/~jakintosh/context/internal/api"
	"git.sr.ht/~jakintosh/context/internal/service"
)

func main() {
	// Initialize DAG service
	dag := service.New()
	handler := api.NewHandler(dag)

	// Root endpoint
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.WriteHeader(http.StatusOK)
			return
		}
		http.NotFound(w, r)
	})

	// Tag endpoints
	http.HandleFunc("/tags", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handler.GetTags(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/tags/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handler.GetTag(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Context endpoints
	http.HandleFunc("/context/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.Trim(r.URL.Path, "/")
		parts := strings.Split(path, "/")

		if len(parts) < 2 {
			http.Error(w, "node ID required", http.StatusBadRequest)
			return
		}

		// Check if this is a tag operation
		if len(parts) >= 3 && parts[2] == "tag" {
			switch r.Method {
			case http.MethodPut:
				handler.PutTag(w, r)
			case http.MethodDelete:
				handler.DeleteTag(w, r)
			default:
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		// Otherwise it's a context operation
		switch r.Method {
		case http.MethodGet:
			handler.GetContext(w, r)
		case http.MethodPost:
			handler.PostContext(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
