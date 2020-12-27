package main

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

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
	command  string
	executed bool
	success  bool
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

type TaskCommand struct {
	taskID  string
	command string
}

// Bot installed on a warehouse roof
type Bot struct {
	mu               sync.RWMutex
	tasks            map[string][]Task
	state            RobotState
	movementDuration time.Duration
	taskCommand      chan TaskCommand

	States chan RobotState
	Errors chan error
}

// NewBot instantiates a bot on a specified location on the roof
func NewBot(x uint, y uint) Bot {
	return Bot{
		tasks:            map[string][]Task{},
		state:            RobotState{X: x, Y: y},
		movementDuration: time.Second,
		taskCommand:      make(chan TaskCommand),
		States:           make(chan RobotState),
		Errors:           make(chan error)}
}

// RunRobot runs the robot to process incoming commands
func (b Bot) RunRobot() {
	for {
		select {
		case tc := <-b.taskCommand:
			log.Printf("Processing command: \"%s\"", tc)
			b.putTask(tc.taskID, Task{tc.command, false, false})
			updatedState, err := b.getUpdatedState(tc.command)
			if err != nil {
				go func() { b.Errors <- err }() // independent consumer can consume errors
			} else {
				log.Printf("New state: %v", updatedState)
				log.Printf("Task states: %v", b.tasks)
				b.state = updatedState
				go func() { b.States <- updatedState }() // independent consumer can consume new states
			}
		case err := <-b.Errors:
			log.Printf("Error: %s", err.Error())
		}
	}
}

func getCommandSequence(commands string) ([]string, error) {
	// Check for empty string
	trimmedCommands := strings.Trim(commands, " ")
	if trimmedCommands == "" {
		return []string{}, fmt.Errorf("Failed to execute empty commands - \"%s\"", commands)
	}

	// Check for invalid command types
	commandSeq := strings.Split(trimmedCommands, " ")
	for _, command := range commandSeq {
		if !strings.Contains("NEWS", command) {
			return []string{}, fmt.Errorf("Invalid command %s, command can only be one of 'N', 'S', 'E' or 'W'", command)
		}
	}

	// TODO check for string with multiple whitespaces

	return commandSeq, nil
}

// EnqueueTask TODO
func (b Bot) EnqueueTask(commands string) (taskID string, position chan RobotState, err chan error) {
	log.Printf("Processing commands: \"%s\"", commands)

	// TODO generate taskID
	taskID = "1"
	position = b.States
	err = b.Errors

	// Asynchronously push commands to channel in-order
	commandSeq, _ := getCommandSequence(commands)
	go func() {
		for _, command := range commandSeq {
			b.taskCommand <- TaskCommand{taskID, command}
		}
	}()

	return
}

func (b Bot) putTask(id string, task Task) {
	b.mu.Lock()
	defer b.mu.Unlock()
	taskList, ok := b.tasks[id]
	if !ok {
		b.tasks[id] = []Task{task}
		log.Println("new")
		return
	}
	b.tasks[id] = append(taskList, task)
}

func (b Bot) getUpdatedState(command string) (RobotState, error) {
	switch command {
	case "N":
		if b.state.Y+1 > 9 {
			return RobotState{}, fmt.Errorf("Command %s exceeds warehouse dimensions", command)
		}
		return RobotState{b.state.X, b.state.Y + 1, b.state.HasCrate}, nil
	case "S":
		if int(b.state.Y)-1 < 0 {
			return RobotState{}, fmt.Errorf("Command %s exceeds warehouse dimensions", command)
		}
		return RobotState{b.state.X, b.state.Y - 1, b.state.HasCrate}, nil
	case "E":
		if b.state.X+1 > 9 {
			return RobotState{}, fmt.Errorf("Command %s exceeds warehouse dimensions", command)
		}
		return RobotState{b.state.X + 1, b.state.Y, b.state.HasCrate}, nil
	case "W":
		if int(b.state.X)-1 < 0 {
			return RobotState{}, fmt.Errorf("Command %s exceeds warehouse dimensions", command)
		}
		return RobotState{b.state.X - 1, b.state.Y, b.state.HasCrate}, nil
	}
	return RobotState{}, fmt.Errorf("Invalid command %s", command)
}

// CancelTask TODO
func (b Bot) CancelTask(taskID string) error {
	return errors.New("as")
}

// CurrentState returns the latest state of the robot
func (b Bot) CurrentState() RobotState {
	return b.state
}
