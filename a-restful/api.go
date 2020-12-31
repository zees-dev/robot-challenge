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

	// static file server for frontend - NOTE: go1.6 can embed files directory into binary
	log.Println(`serving frontend at "/"...`)
	router.Handle("/", http.FileServer(http.Dir("./public")))

	// serve swagger ui - NOTE: go1.6 can embed files directory into binary
	log.Println(`serving open api spec at "/swaggerui/"...`)
	swaggerui := http.StripPrefix("/swaggerui/", http.FileServer(http.Dir("./swaggerui/")))
	router.PathPrefix("/swaggerui/").Handler(swaggerui)

	// server health endpoint
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
		id, ok := vars["id"]
		if !ok {
			http.Error(w, "missing request id", http.StatusBadRequest)
			return
		}

		err := robot.CancelTask(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		// TODO return 400 for scenario where task has already been executed

		w.WriteHeader(http.StatusNoContent)
	}).Methods("DELETE")

	// CHALLENGE
	// HTTP2 SSE - realtime unidirectional communication
	router.HandleFunc("/api/v1/state/subscribe", func(w http.ResponseWriter, r *http.Request) {
		log.Println("established handshake with client...")

		// ensure writer supports streaming
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
			return
		}

		// set SSE headers
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// TODO - gracefully handle client disconnections
		// - this is just a POC to demonstrate real-time updates to single client using SSE)

		// enqueue empty task to hook into state and error channels
		_, stateCh, errorsCh := robot.EnqueueTask("")

		// Event stream format/spec: https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events#Event_stream_format
		for {
			select {
			case state := <-stateCh:
				log.Printf("SSE recieving - robot state %v", state)
				fmt.Fprint(w, "event: robotstate\n")
				fmt.Fprintf(w, `data: {"x": %d, "y": %d}%s`, state.X, state.Y, "\n")
				fmt.Fprint(w, "\n")
				flusher.Flush()
			case err := <-errorsCh:
				log.Printf("SSE recieving - error %v", err)
				fmt.Fprint(w, "event: roboterror\n")
				fmt.Fprintf(w, `data: %s%s`, err.Error(), "\n")
				fmt.Fprint(w, "\n")
				flusher.Flush()
			case <-r.Context().Done():
				return
			}
		}
	}).Methods("GET")

	return router
}
