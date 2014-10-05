// This package implements a simple HTTP server providing a REST API to a task handler.
//
// It provides four methods:
//
// 	GET    /task/          Retrieves all the tasks.
// 	POST   /task/          Creates a new task given a title.
// 	GET    /task/{taskID}  Retrieves the task with the given id.
// 	PUT    /task/{taskID}  Updates the task with the given id.
//
// Every method below gives more information about every API call, its parameters, and its results.
package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/timbogit/todo/task"
)

var tasks = task.NewTaskManager()

const PathPrefix = "/task/"

func RegisterHandlers() {
	r := mux.NewRouter()
	r.HandleFunc(PathPrefix, errorHandler(ListTasks)).Methods("GET")
	r.HandleFunc(PathPrefix, errorHandler(NewTask)).Methods("POST")
	r.HandleFunc(PathPrefix, errorHandler(ReplaceTasks)).Methods("PUT")
	r.HandleFunc(PathPrefix+"{id}", errorHandler(GetTask)).Methods("GET")
	r.HandleFunc(PathPrefix+"{id}", errorHandler(UpdateTask)).Methods("PUT")
	http.Handle(PathPrefix, r)
}

// badRequest is handled by setting the status code in the reply to StatusBadRequest.
type badRequest struct{ error }

// notFound is handled by setting the status code in the reply to StatusNotFound.
type notFound struct{ error }

// errorHandler wraps a function returning an error by handling the error and returning a http.Handler.
// If the error is of the one of the types defined above, it is handled as described for every type.
// If the error is of another type, it is considered as an internal error and its message is logged.
func errorHandler(f func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := f(w, r)
		if err == nil {
			log.Println(r)
			return
		}
		switch err.(type) {
		case badRequest:
			log.Println(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
		case notFound:
			log.Println(err)
			http.Error(w, "task not found", http.StatusNotFound)
		default:
			log.Println(err)
			http.Error(w, "oops", http.StatusInternalServerError)
		}
	}
}

// ListTask handles GET requests on /task.
// There's no parameters and it returns an object with a Tasks field containing a list of tasks.
//
// Example:
//
//   req: GET /task/
//   res: 200 {"Tasks": [
//          {"id": 1, "title": "Learn Go", "completed": false},
//          {"id": 2, "title": "Buy bread", "completed": true}
//        ]}
func ListTasks(w http.ResponseWriter, r *http.Request) error {
	res := struct{ Tasks []*task.Task }{tasks.All()}
	return json.NewEncoder(w).Encode(res)
}

// NewTask handles POST requests on /task.
// The request body must contain a JSON object with a Title field.
// The status code of the response is used to indicate any error.
//
// Examples:
//
//   req: POST /task/ {"title": ""}
//   res: 400 empty title
//
//   req: POST /task/ {"title": "Buy bread"}
//   res: 200
func NewTask(w http.ResponseWriter, r *http.Request) error {
	req := struct{ Title string }{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return badRequest{err}
	}
	t, err := task.NewTask(req.Title)
	if err != nil {
		return badRequest{err}
	}
	return tasks.Save(t)
}

// parseID obtains the id variable from the given request url,
// parses the obtained text and returns the result.
func parseID(r *http.Request) (int64, error) {
	txt, ok := mux.Vars(r)["id"]
	if !ok {
		return 0, fmt.Errorf("task id not found")
	}
	return strconv.ParseInt(txt, 10, 0)
}

// GetTask handles GET requsts to /task/{taskID}.
// There's no parameters and it returns a JSON encoded task.
//
// Examples:
//
//   req: GET /task/1
//   res: 200 {"id": 1, "title": "Buy bread", "completed": true}
//
//   req: GET /task/42
//   res: 404 task not found
func GetTask(w http.ResponseWriter, r *http.Request) error {
	id, err := parseID(r)
	log.Println("Task is ", id)
	if err != nil {
		return badRequest{err}
	}
	t, ok := tasks.Find(id)
	log.Println("Found", ok)

	if !ok {
		return notFound{}
	}
	return json.NewEncoder(w).Encode(t)
}

// UpdateTask handles PUT requests to /task/{taskID}.
// The request body must contain a JSON encoded task.
//
// Example:
//
//   req: PUT /task/1 {"id": 1, "title": "Learn Go", "completed": true}
//   res: 200
//
//   req: PUT /task/2 {"id": 2, "title": "Learn Go", "completed": true}
//   res: 400 inconsistent task IDs
func UpdateTask(w http.ResponseWriter, r *http.Request) error {
	id, err := parseID(r)
	if err != nil {
		return badRequest{err}
	}
	var t task.Task
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		return badRequest{err}
	}
	if t.ID != id {
		return badRequest{fmt.Errorf("inconsistent task IDs")}
	}
	if _, ok := tasks.Find(id); !ok {
		return notFound{}
	}
	return tasks.Save(&t)
}

// ReplaceTasks handles PUT requests to /task/
// The request body must contain a JSON encoded list of tasks.
//
// Example:
//
//   req: PUT /task/ {"Tasks": [
//          {"id":1,"title":"learn go","completed":true},
//          {"id":2,"title":"PROFIT!","completed":false}
//        ]}
//   res: 200 {"Tasks": [
//          {"id":1,"title":"learn go","completed":true},
//          {"id":2,"title":"PROFIT!","completed":false}
//        ]}

func ReplaceTasks(w http.ResponseWriter, r *http.Request) error {
	req := struct{ Tasks []*task.Task }{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return badRequest{err}
	}
	res := tasks.ReplaceAll(req.Tasks)
	return json.NewEncoder(w).Encode(res)
}
