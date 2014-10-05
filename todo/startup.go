// A stand-alone HTTP server providing a web UI for the todo list.
package main

import (
	"net/http"

	"github.com/timbogit/todo/server"
)

func main() {
	server.RegisterHandlers()
	http.Handle("/", http.FileServer(http.Dir("public")))
	http.ListenAndServe(":8080", nil)
}
