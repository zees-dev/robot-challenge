package main

import (
	"fmt"
)

// Warehouse is a 10x10 grid which contains robots
type Warehouse interface {
	Robots() []Robot
}

// Robot navigate a warehouse using `N`, `S`, `E`, `W` commands
type Robot interface {
	EnqueueTask(commands string) (taskID string, position chan RobotState, err chan error)
	CancelTask(taskID string) error
	CurrentState() RobotState
}

// RobotState is current of a singular robot on the warehouse
type RobotState struct {
	X        uint
	Y        uint
	HasCrate bool
}

func main() {
	fmt.Println("hello world")
}
