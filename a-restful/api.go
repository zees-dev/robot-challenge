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

type CreateWarehouse struct {
	ID     uint64 `json:"id"`
	Width  uint32 `json:"width"`
	Height uint32 `json:"height"`
}

type CreateBot struct {
	ID          uint64 `json:"id"`
	WarehouseID uint64 `json:"warehouseId"`
	X           uint   `json:"x"`
	Y           uint   `json:"y"`
}

type UpdateBot struct {
	Commands string `json:"commands"`
}

// BodyToCreateWarehouse marshals request body to CreateWarehouse
func BodyToCreateWarehouse(reqBody io.Reader) (CreateWarehouse, error) {
	var obj CreateWarehouse
	err := json.NewDecoder(reqBody).Decode(&obj)
	if err != nil {
		log.Printf("Error converting body to CreateWarehouse: %v", err)
		return CreateWarehouse{}, errors.New("failed to read request body")
	}
	return obj, nil
}

// BodyToBot marshals request body to CreateBot
func BodyToBot(reqBody io.Reader) (CreateBot, error) {
	var obj CreateBot
	err := json.NewDecoder(reqBody).Decode(&obj)
	if err != nil {
		log.Printf("Error converting body to CreateBot: %v", err)
		return CreateBot{}, errors.New("failed to read request body")
	}
	return obj, nil
}

// BodyToUpdateBot marshals request body to UpdateBot
func BodyToUpdateBot(reqBody io.Reader) (UpdateBot, error) {
	var obj UpdateBot
	err := json.NewDecoder(reqBody).Decode(&obj)
	if err != nil {
		log.Printf("Error converting body to UpdateBot: %v", err)
		return UpdateBot{}, errors.New("failed to read request body")
	}
	return obj, nil
}

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
