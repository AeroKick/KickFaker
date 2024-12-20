package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"strings"
)

//go:embed web/dist
var content embed.FS

func main() {
	go manager.run()

	// CORS middleware
	corsMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}

	// Create a new mux for handling routes
	mux := http.NewServeMux()

	// Handle WebSocket connections
	mux.HandleFunc("/ws", handleWebSocket)

	// Serve static files from the embedded filesystem
	fsys, err := fs.Sub(content, "web/dist")
	if err != nil {
		log.Fatal(err)
	}

	// Create a file server handler
	fileServer := http.FileServer(http.FS(fsys))

	// Handle all other routes
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Serve index.html for all routes except /ws and static files
		if strings.HasPrefix(r.URL.Path, "/ws") {
			http.NotFound(w, r)
			return
		}

		// Check if the file exists
		if _, err := fs.Stat(fsys, strings.TrimPrefix(r.URL.Path, "/")); err != nil {
			// If file doesn't exist, serve index.html
			r.URL.Path = "/"
		}

		fileServer.ServeHTTP(w, r)
	})

	// Wrap the mux with CORS middleware
	handler := corsMiddleware(mux)

	log.Println("Server starting on :4400")
	if err := http.ListenAndServe(":4400", handler); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
