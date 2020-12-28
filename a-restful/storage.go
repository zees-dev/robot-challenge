package main

import (
	"fmt"
	"sync"
)

// Repository contains signature which a storage/persistent layer must implement
type Repository interface {
	GetTask(id string) (Task, error)
	CreateTask(ct Task) error
	UpdateTask(ut Task) error
}

// DB is a struct which stores robot tasks in-memory
type DB struct {
	mu    sync.RWMutex // RW mutex to allow multiple readers but single writer
	tasks []Task
}

// NewDB instantiates empty database of robot tasks
func NewDB() *DB {
	return &DB{}
}

// GetTask gets structure from in-memory DB by ID in a concurrent-safe way
func (db *DB) GetTask(id string) (Task, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	for _, t := range db.tasks {
		if t.id == id {
			return t, nil
		}
	}
	return Task{}, fmt.Errorf("Task with ID: %s not found", id)
}

// CreateTask creates task in in-memory DB by ID in a concurrent-safe way
func (db *DB) CreateTask(ct Task) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	for _, t := range db.tasks {
		if t.id == ct.id {
			return fmt.Errorf("Task with ID: %s already exists", t.id)
		}
	}
	db.tasks = append(db.tasks, ct)
	return nil
}

// UpdateTask updates task in in-memory DB in a concurrent-safe way
func (db *DB) UpdateTask(ut Task) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	for i, t := range db.tasks {
		if t.id == ut.id {
			db.tasks[i] = ut
			return nil
		}
	}
	return fmt.Errorf("Task with ID: %s not found", ut.id)
}
