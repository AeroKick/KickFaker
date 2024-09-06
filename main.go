package main

import (
	_ "embed"
	"fmt"
	"log"
	"net/http"
)

func main() {
	go manager.run()
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/app/", handleWebSocket)

	fmt.Println("Server is running on http://localhost:4400. WS Server is running on ws://localhost:4400/app")
	log.Fatal(http.ListenAndServe(":4400", nil))
}

//go:embed web/index.html
var html []byte

func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Write(html)
}
