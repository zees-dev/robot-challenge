package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// UpdateBot is request body to update robot state
type UpdateBot struct {
	Commands string `json:"commands"`
}

// BodyToUpdateBot marshals request body to UpdateBot struct
func BodyToUpdateBot(reqBody io.Reader) (UpdateBot, error) {
	var obj UpdateBot
	err := json.NewDecoder(reqBody).Decode(&obj)
	if err != nil {
		log.Printf("Error converting body to UpdateBot: %v", err)
		return UpdateBot{}, errors.New("failed to read request body")
	}
	return obj, nil
}

// robotServer is the Restful API server exposed by robot which enables ground control station to communicate with it
func robotServer(robot Bot) {
	router := mux.NewRouter()

	// Robot movement
	router.HandleFunc("/move", func(w http.ResponseWriter, r *http.Request) {
		// TODO use request context

		body, err := BodyToUpdateBot(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = validateCommandSequence(body.Commands)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		taskID, _, _ := robot.EnqueueTask(body.Commands)

		fmt.Fprintf(w, "%s", taskID)
	}).Methods("PUT")

	// router.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
	// 	// TODO use request context

	// 	tasks := robot.getTask()
	// 	fmt.Fprintf(w, "%v", tasks)
	// }).Methods("GET")

	// GetTask
	router.HandleFunc("/task/{id}", func(w http.ResponseWriter, r *http.Request) {
		// TODO use request context
		vars := mux.Vars(r)
		id, ok := vars["id"]
		if !ok {
			http.Error(w, "missing request id", http.StatusBadRequest)
			return
		}

		task, err := robot.getTask(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		fmt.Fprintf(w, "%v", task)
	}).Methods("GET")

	// Cancel Task
	router.HandleFunc("/task/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		err := robot.CancelTask(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}).Methods("DELETE")

	log.Println("Starting admin server on :8000")
	err := http.ListenAndServe(":8000", router)
	log.Fatal(err)
}
