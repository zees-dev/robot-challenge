package main

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	uuid "github.com/satori/go.uuid"
)

// Warehouse is the structure in which robots operate
// - robots operate on a 10x10 grid on the roof of the warehouse
type Warehouse interface {
	Robots() []Robot
}

// Robot navigate a warehouse using `N`, `S`, `E`, `W` commands
type Robot interface {
	EnqueueTask(commands string) (taskID string, position chan RobotState, err chan error)
	CancelTask(taskID string) error
	CurrentState() RobotState
}

// RobotState is current state of a singular robot on the warehouse roof
type RobotState struct {
	X        uint
	Y        uint
	HasCrate bool
}

// Task is used to identify whether robot has successfully completed a sequence of commands
type Task struct {
	command   string
	executed  bool
	success   bool
	cancelled bool
}

// Building is the a concept building/site which implements Warehouse interface
type Building struct {
	bots   []Robot
	Width  uint32
	Height uint32
}

// NewBuilding building in which the robot operates
func NewBuilding(id uint64) Building {
	return Building{bots: []Robot{}, Width: 10, Height: 10}
}

// RegisterRobot registers a new robot into building/warehouse
func (b Building) RegisterRobot(r Robot) []Robot {
	return append(b.bots, r)
}

// Robots implements Warehouse
func (b Building) Robots() []Robot {
	return b.bots
}

// Bot installed on a warehouse roof
// * implements robot interface
type Bot struct {
	mu               sync.RWMutex
	taskMap          map[string]Task
	state            RobotState
	movementDuration time.Duration
	tasks            chan string

	States chan RobotState
	Errors chan error
}

// NewBot instantiates a bot on a specified location on the roof
func NewBot(x uint, y uint) Bot {
	return Bot{
		taskMap: map[string]Task{},
		state:   RobotState{X: x, Y: y},
		tasks:   make(chan string),
		States:  make(chan RobotState),
		Errors:  make(chan error)}
}

// RunRobot runs the robot to process incoming commands
func (b *Bot) RunRobot() {
	for {
		select {
		case taskID := <-b.tasks:
			taskToProcess, err := b.getTask(taskID)
			if err != nil {
				log.Printf("Task %s cannot be processed - not found", taskID)
				go func() { b.Errors <- err }()
			} else {
				log.Printf("Processing command: \"%s\"", taskToProcess.command)
				updatedState, err := b.getUpdatedState(taskToProcess.command)
				taskToProcess.executed = true
				if err != nil {
					b.putTask(taskID, taskToProcess)
					go func() { b.Errors <- err }() // independent consumer can consume errors
				} else {
					log.Printf("Updating robot to new state: %v", updatedState)
					b.UpdateCurrentState(updatedState)
					go func() { b.States <- updatedState }() // independent consumer can consume state changes

					log.Printf("Task states: %v", b.tasks) // TODO remove
					taskToProcess.success = true
					b.putTask(taskID, taskToProcess)
				}
			}
		case err := <-b.Errors: // Example of consumer consuming errors
			log.Printf("Error: %s", err.Error())
		}
	}
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

// EnqueueTask queues a task on the `taskCommand` bot channel to be processed by `RunRobot` function
// * implements robot
func (b Bot) EnqueueTask(commands string) (taskID string, position chan RobotState, err chan error) {
	log.Printf("Processing commands: \"%s\"", commands)

	taskID = uuid.NewV4().String()
	position = b.States
	err = b.Errors

	b.putTask(taskID, Task{commands, false, false, false})
	b.tasks <- taskID

	return
}

// putTask stores the task on a map
func (b Bot) putTask(id string, task Task) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.taskMap[id] = task
}

// getTask retrieves a task from the map - base on its id
func (b Bot) getTask(id string) (Task, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	task, ok := b.taskMap[id]
	if !ok {
		return Task{}, fmt.Errorf("Unable to find task %s", id)
	}
	return task, nil
}

// getUpdatedState translates a sequence of space delimited movement commands to a final RobotState
func (b Bot) getUpdatedState(commands string) (RobotState, error) {
	finalState := b.state
	for _, command := range commands {
		switch string(command) {
		case "N":
			if finalState.Y++; finalState.Y > 9 {
				return RobotState{}, fmt.Errorf("Command %s exceeds warehouse dimensions", string(command))
			}
		case "S":
			if int(finalState.Y)-1 < 0 {
				return RobotState{}, fmt.Errorf("Command %s exceeds warehouse dimensions", string(command))
			}
			finalState.Y--
		case "E":
			if finalState.X++; finalState.X > 9 {
				return RobotState{}, fmt.Errorf("Command %s exceeds warehouse dimensions", string(command))
			}
		case "W":
			if int(finalState.X)-1 < 0 {
				return RobotState{}, fmt.Errorf("Command %s exceeds warehouse dimensions", string(command))
			}
			finalState.X--
		}
	}
	return finalState, nil
}

// CancelTask sets an existing task on the map to be cancelled
// * implements robot
func (b Bot) CancelTask(taskID string) error {
	task, err := b.getTask(taskID)
	if err != nil {
		return err
	}
	task.cancelled = true
	b.putTask(taskID, task)
	return nil
}

// UpdateCurrentState current state concurrent-safe
func (b *Bot) UpdateCurrentState(rs RobotState) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.state = rs
}

// CurrentState returns the latest state of the robot
// * implements robot
func (b Bot) CurrentState() RobotState {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.state
}
