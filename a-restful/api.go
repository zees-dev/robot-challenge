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
	trimmedCommands := strings.TrimSpace(commands)
	if trimmedCommands == "" {
		return errors.New("failed to execute empty commands")
	}

	// check for multiple whitespaces
	if strings.Contains(commands, "  ") {
		return fmt.Errorf(`invalid command '%s'; command  cannot contain multiple whitespaces`, commands)
	}

	// Check for invalid command types
	commandSeq := strings.Split(trimmedCommands, " ")
	for _, command := range commandSeq {
		if !strings.Contains("NEWS", command) {
			return fmt.Errorf(`invalid command '%s', command can only be one of 'N', 'S', 'E' or 'W'`, command)
		}
	}

	return nil
}

// RobotAPIServer is the Restful API server exposed by robot which enables ground control station to communicate with it
// Note: This could require `Robot` instead of `Bot` - but `Robot` does not have the `GetTask` method - which is a requirement...
// - requirement: "Create a RESTful API to report the command series's execution status"
func RobotAPIServer(robot *Bot) http.Handler {
	router := mux.NewRouter()

	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"healthy"}`))
	}).Methods("GET")

	// Robot state
	router.HandleFunc("/api/v1/state", func(w http.ResponseWriter, r *http.Request) {
		// TODO use request context
		w.Header().Set("Content-Type", "application/json")

		state := robot.CurrentState()
		fmt.Fprintf(w, `{"x": %d, "y": %d}`, state.X, state.Y)
	}).Methods("GET")

	// Robot movement
	router.HandleFunc("/api/v1/state", func(w http.ResponseWriter, r *http.Request) {
		// TODO use request context
		w.Header().Set("Content-Type", "application/json")

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

		fmt.Fprintf(w, `{"taskID": "%s"}`, taskID)
	}).Methods("PUT")

	// GetTask by id
	router.HandleFunc("/api/v1/task/{id}", func(w http.ResponseWriter, r *http.Request) {
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

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"task":{"id":"%s","command":"%s","executed":%t,"cancelled":%t,"success":%t}}`, task.id, task.command, task.executed, task.cancelled, task.success)
	}).Methods("GET")

	// Cancel Task by id
	router.HandleFunc("/api/v1/task/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		err := robot.CancelTask(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}).Methods("DELETE")

	return router
}
