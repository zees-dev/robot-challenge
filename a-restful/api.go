package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

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

// validateCommandSequence will validate string delimited movement input
// only `N`, `S`, `E` and `W` characters are allowed within space-delimited string
func validateCommandSequence(commands string) error {
	// Check for empty string
	trimmedCommands := strings.Trim(commands, " ")
	if trimmedCommands == "" {
		return fmt.Errorf("Failed to execute empty commands - \"%s\"", commands)
	}

	// Check for invalid command types
	commandSeq := strings.Split(trimmedCommands, " ")
	for _, command := range commandSeq {
		if !strings.Contains("NEWS", command) {
			return fmt.Errorf("Invalid command %s, command can only be one of 'N', 'S', 'E' or 'W'", command)
		}
	}

	// TODO check for string with multiple whitespaces

	return nil
}

// robotAPIServer is the Restful API server exposed by robot which enables ground control station to communicate with it
func robotAPIServer(robot Bot) {
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

	// GetTask by id
	router.HandleFunc("/task/{id}", func(w http.ResponseWriter, r *http.Request) {
		// TODO use request context
		vars := mux.Vars(r)
		id, ok := vars["id"]
		if !ok {
			http.Error(w, "missing request id", http.StatusBadRequest)
			return
		}

		task, err := robot.repository.GetTask(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		fmt.Fprintf(w, "%v", task)
	}).Methods("GET")

	// Cancel Task by id
	router.HandleFunc("/task/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		err := robot.CancelTask(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}).Methods("DELETE")

	log.Println("Starting admin server on :8000...")
	err := http.ListenAndServe(":8000", router)
	log.Fatal(err)
}