package main

import (
	"fmt"
	"log"
	"sync"

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
	id        string
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
	mu         sync.RWMutex
	repository Repository
	state      RobotState
	tasks      chan string

	States chan RobotState
	Errors chan error
}

// NewBot instantiates a bot on a specified location on the roof
func NewBot(x uint, y uint, repository Repository) Bot {
	return Bot{
		repository: repository,
		state:      RobotState{X: x, Y: y},
		tasks:      make(chan string),
		States:     make(chan RobotState),
		Errors:     make(chan error)}
}

// RunRobot runs the robot to process incoming commands
func (b *Bot) RunRobot() {
	log.Println("Running robot, listening to operations...")
	for {
		select {
		case taskID := <-b.tasks:
			// Wrap up in func to increase readability as we cannot break out of for..select
			func() {
				taskToProcess, err := b.repository.GetTask(taskID)
				if err != nil {
					log.Printf("Task %s cannot be processed - not found", taskID)
					go func() { b.Errors <- err }()
					return
				}

				if taskToProcess.cancelled {
					log.Printf("Task %s has been cancelled", taskID)
					return
				}

				log.Printf(`Processing task "%s": "%s"`, taskID, taskToProcess.command)
				updatedState, err := b.getUpdatedState(taskToProcess.command)
				taskToProcess.executed = true
				if err != nil {
					log.Printf("error: %s", err)
					b.repository.UpdateTask(taskToProcess)
					go func() { b.Errors <- err }() // independent consumer can consume errors
					return
				}

				log.Printf("Updating robot to new state: %v", updatedState)
				err = b.UpdateCurrentState(updatedState)
				if err != nil {
					log.Printf("failed to update robot to new state: %v", updatedState)
					go func() { b.Errors <- err }() // independent consumer can consume errors
					return
				}

				go func() { b.States <- updatedState }() // independent consumer can consume state changes
				taskToProcess.success = true
				b.repository.UpdateTask(taskToProcess)
			}()
		}
	}
}

// EnqueueTask queues a task on the `taskCommand` bot channel to be processed by `RunRobot` function
// * implements robot
func (b Bot) EnqueueTask(commands string) (taskID string, position chan RobotState, err chan error) {
	log.Printf("Queueing commands: \"%s\"", commands)

	taskID = uuid.NewV4().String()
	position = b.States
	err = b.Errors

	b.repository.CreateTask(Task{taskID, commands, false, false, false})
	b.tasks <- taskID

	return
}

// getUpdatedState translates a sequence of space delimited movement commands to a final RobotState
func (b Bot) getUpdatedState(commands string) (RobotState, error) {
	finalState := b.state
	for _, command := range commands {
		switch string(command) {
		case "N":
			if finalState.Y++; finalState.Y > 9 {
				return RobotState{}, fmt.Errorf("command `%s` exceeds warehouse dimensions", string(command))
			}
		case "S":
			if int(finalState.Y)-1 < 0 {
				return RobotState{}, fmt.Errorf("command `%s` exceeds warehouse dimensions", string(command))
			}
			finalState.Y--
		case "E":
			if finalState.X++; finalState.X > 9 {
				return RobotState{}, fmt.Errorf("command `%s` exceeds warehouse dimensions", string(command))
			}
		case "W":
			if int(finalState.X)-1 < 0 {
				return RobotState{}, fmt.Errorf("command `%s` exceeds warehouse dimensions", string(command))
			}
			finalState.X--
		}
	}
	return finalState, nil
}

// CancelTask sets an existing task on the map to be cancelled
// * implements robot
func (b Bot) CancelTask(taskID string) error {
	task, err := b.repository.GetTask(taskID)
	if err != nil {
		return err
	}
	if task.executed {
		return fmt.Errorf("task %s has already been executed", taskID)
	}

	task.cancelled = true
	b.repository.UpdateTask(task)
	return nil
}

// UpdateCurrentState current state concurrent-safe way; additionally this method ensures robotstate lies within warehouse dimensions
func (b *Bot) UpdateCurrentState(rs RobotState) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if rs.X < 0 || rs.X > 9 {
		return fmt.Errorf("Robot state X position (%d, y) exceeds warehouse dimensions", rs.X)
	}
	if rs.Y < 0 || rs.Y > 9 {
		return fmt.Errorf("Robot state Y position (x, %d) exceeds warehouse dimensions", rs.Y)
	}
	b.state = rs
	return nil
}

// CurrentState returns the latest state of the robot
// * implements robot
func (b Bot) CurrentState() RobotState {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.state
}
