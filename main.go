package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io/ioutil"
	"net/http"
	"sync"
)

type HealthResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}
type Todo struct {
	Id        string `json:"id"`
	Task      string `json:"task"`
	Completed bool   `json:"completed"`
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	health := HealthResponse{
		Status:  "OK",
		Message: "API HEALTH",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}
func main() {
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/todos", todoHandler)
	http.HandleFunc("/todos/", todoByIdHandler)
	fmt.Println("App is running at port 3000")
	err := http.ListenAndServe(":3000", nil)

	if err != nil {
		fmt.Println("error", err)
	}
}

var (
	todos     []Todo
	todoMutex sync.Mutex // only one goroutine at a time
)

func generatenewId() string {
	return uuid.New().String()
}
func todoHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.Method)
	switch r.Method {
	case "GET":
		w.Header().Set("content-type", "application/json")
		json.NewEncoder(w).Encode(todos)

	case "POST":
		var newtodo Todo
		var body, err = ioutil.ReadAll(r.Body)
		if err != nil {
			//w.WriteHeader(http.StatusBadRequest)
			http.Error(w, "unable to read from request body", http.StatusBadRequest)
			return
		}
		//fmt.Println(body)
		err = json.Unmarshal(body, &newtodo)
		if err != nil || newtodo.Task == "" {
			http.Error(w, "no inputs found", http.StatusBadRequest)
			return
		}
		newtodo.Id = generatenewId()
		todoMutex.Lock()
		todos = append(todos, newtodo)
		todoMutex.Unlock()
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(newtodo)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func todoByIdHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/todos/"):]
	todoMutex.Lock()
	defer todoMutex.Unlock()
	for i, todo := range todos {
		if todo.Id == id {
			switch r.Method {
			case "GET":
				w.Header().Set("content-type", "application/json")
				json.NewEncoder(w).Encode(todo)
			case "PUT":
				var updatedTodo Todo
				body, err := ioutil.ReadAll(r.Body)
				if err != nil {
					http.Error(w, "unable to read from request body", http.StatusBadRequest)
					return
				}
				err = json.Unmarshal(body, &updatedTodo)
				if err != nil || updatedTodo.Task == "" {
					http.Error(w, "Invalid input", http.StatusBadRequest)
					return
				}
				todos[i].Task = updatedTodo.Task
				todos[i].Completed = updatedTodo.Completed
				json.NewEncoder(w).Encode(todos[i])
			case "DELETE":
				deletedTodoId := todo.Id
				todos = append(todos[:i], todos[i+1:]...)
				w.Header().Set("content-type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]string{"message ": "todo with id " + deletedTodoId + "is deleted"})
			default:
			}
		}
	}
}
