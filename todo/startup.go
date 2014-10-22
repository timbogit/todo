// A stand-alone HTTP server providing a web UI for the todo list.
package main

import (
	"net/http"
	"os"

	"github.com/timbogit/todo/server"
)

func main() {
	server.RegisterHandlers()
	http.Handle("/", http.FileServer(http.Dir("todo/public")))
	http.ListenAndServe(":" + os.Getenv("PORT"), nil)
}
