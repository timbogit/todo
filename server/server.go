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
	jwt "github.com/dgrijalva/jwt-go"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/timbogit/todo/task"
)

const (
	PrivKeyPath = "../keys/demo.rsa"     // openssl genrsa -out demo.rsa 1024
	PubKeyPath  = "../keys/demo.rsa.pub" // openssl rsa -in demo.rsa -pubout > demo.rsa.pub
	PathPrefix  = "/task/"
	AuthPrefix  = "/login/"
)

var (
	tasks     = task.NewTaskManager()
	signKey   []byte
	verifyKey []byte
)

// read the key files before starting http handlers
func init() {
	var err error

	signKey, err = ioutil.ReadFile(PrivKeyPath)
	if err != nil {
		log.Fatal("Error reading private key")
		return
	}

	verifyKey, err = ioutil.ReadFile(PubKeyPath)
	if err != nil {
		log.Fatal("Error reading private key")
		return
	}
}

func RegisterHandlers() {
	r := mux.NewRouter()
	r.HandleFunc(PathPrefix, errorHandler(authHandler(ListTasks))).Methods("GET")
	r.HandleFunc(PathPrefix, errorHandler(authHandler(NewTask))).Methods("POST")
	r.HandleFunc(PathPrefix, errorHandler(authHandler(ReplaceTasks))).Methods("PUT")
	r.HandleFunc(PathPrefix+"{id}", errorHandler(authHandler(GetTask))).Methods("GET")
	r.HandleFunc(PathPrefix+"{id}", errorHandler(authHandler(UpdateTask))).Methods("PUT")

	http.Handle(PathPrefix, r)

	r.HandleFunc(AuthPrefix, errorHandler(CreateToken)).Methods("POST")
	http.Handle(AuthPrefix, r)
}

// badRequest is handled by setting the status code in the reply to StatusBadRequest.
type badRequest struct{ error }

// notFound is handled by setting the status code in the reply to StatusNotFound.
type notFound struct{ error }

// unauthorized is for resources that are restricted to requests with valid JWT tokens
type unauthorized struct{ error }

// forbidden is for requests for the auth endpoint if the user didn't provide a correct password
type forbidden struct{ error }

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
		case unauthorized:
			log.Println(err)
			http.Error(w, "access restricted", http.StatusUnauthorized)
		case forbidden:
			log.Println(err)
			http.Error(w, "wrong credentials", http.StatusForbidden)
		default:
			log.Println(err)
			http.Error(w, "oops", http.StatusInternalServerError)
		}
	}
}

// authHandler wraps a function returning an error by first checking for the correct JWT token and returning the same function.
// If the JWT token is absent, or empty, it returns a function that simply returns the `unauthorized` error.
//
// example:
//   req: POST /login/ {"user": "test", "password":"known"}
//   res: 200 {
//              "token":"f00ba7"
//            }
//   req: GET /task/1, header "Authorization: Bearer f00ba7"
//   res: 200 {"id": 1, "title": "Buy bread", "completed": true}
func authHandler(f func(w http.ResponseWriter, r *http.Request) error) func(w http.ResponseWriter, r *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		log.Println("Request Headers are:", r.Header)
		token, err := jwt.ParseFromRequest(r, func(token *jwt.Token) (interface{}, error) {
			return verifyKey, nil
		})
		if token != nil && token.Valid {
			return f(w, r)
		} else {
			return unauthorized{err}
		}
	}
}

// CreateToken handles POST requests on /login/.
// It accepts as `user` and `password` parameter, and it returns a JWT token
//
// Example:
//
//   req: POST /login/ {"user": "test", "password":"known"}
//   res: 200 {
//              "token":"eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJBY2Nlc3NUb2tlbiI6ImxldmVsMSIsIkN1c3RvbVVzZXJJbmZvIjp7Ik5hbWUiOiJ0ZXN0IiwiS2luZCI6Imh1bWFuIn0sImV4cCI6MTQxMzExMTEyOH0.TuQL786NjIDLwyhG60lQd-bdmd98gTxUSKaMLXDUyqJQd_MxSuK7wKbDQUSoU4ux5uLHaCPxg0H4-7zm0P0TGYlVd7UaFsxu5VakXihsYO69V_2UsdkJqtMuwGoelqqDuOAAP5vLsdQjaJ7a5KGAlE0647ozosJgEf-Ujp4pX9g"
//            }
func CreateToken(w http.ResponseWriter, r *http.Request) error {
	res := struct {
		Token string `json:"token"`
	}{""}

	req := struct {
		User     string
		Password string
	}{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return badRequest{err}
	}

	user := req.User
	pass := req.Password

	log.Printf("Authenticate: user[%s] pass[%s]\n", user, pass)

	// check values
	if user != "test" || pass != "known" {
		return forbidden{}
	}

	// create a signer for rsa 256
	t := jwt.New(jwt.GetSigningMethod("RS256"))

	// set our claims
	t.Claims["AccessToken"] = "level1"
	t.Claims["CustomUserInfo"] = struct {
		Name string
		Kind string
	}{user, "human"}

	// set the expire time
	// see http://tools.ietf.org/html/draft-ietf-oauth-json-web-token-20#section-4.1.4
	t.Claims["exp"] = time.Now().Add(time.Hour * 10).Unix()
	tokenString, err := t.SignedString(signKey)
	if err != nil {
		log.Printf("Token Signing error: %v\n", err)
		return err
	}

	// store the token string in the response struct and json-ify
	res.Token = tokenString
	return json.NewEncoder(w).Encode(res)
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
	log.Println("Incoming request parsed is: ", req)
	response := struct{ Tasks []*task.Task }{tasks.ReplaceAll(req.Tasks)}
	log.Println("Response objects are: ", response)
	return json.NewEncoder(w).Encode(response)
}
