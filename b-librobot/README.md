# Robot Simulator Library

We wish to create a simulator which mimics the behaviour of our new robots.

The simulation should take the form of a Golang library with associated documentation and tests.

## Library Interface

The simulator should implement the following interfaces (see the included file):

```
type Warehouse interface {
	Robots() []Robot
}

type Robot interface {
	EnqueueTask(commands string) (taskID string, position chan RobotState, err chan error) 

	CancelTask(taskID string) error

	CurrentState() RobotState
}

type RobotState struct {
	X uint
	Y uint
	HasCrate bool
}
```

## Requirements

### Part One

Implement the simulator so that you can create an instance of a simulated `Warehouse`, then add one or more `Robots` to it, and issue instructions to those Robots.

Some notes:
* Only one robot should be able to occupy a location within the warehouse at a time.
* Multiple robots may operate within a single warehouse.
* Multiple warehouses may be simulated at a time.
* Each robot should take one second of real time to perform each command.

Provide documentation and tests to allow users of library to use the simulator and validate its correct operation.

### Part Two

Now, we add a lifting claw to the robot, so that it can move crates in the warehouse.

Add a new interface:

```
type CrateWarehouse interface {
	Warehouse

	AddCrate(x uint, y uint) error
	DelCrate(x uint, y uint) error
}
```

Then extend the valid commands supported by the robot simulator to include the following:
* "G" - If the robot is at a location with a crate, grab it.
* "D" - Drop a carried crate at the robot's current position.

Some notes:
* The robot should only be able to carry one crate at a time.
* A crate may not be dropped at a location where there is already a crate.

Provide tests to validate the correct simulation of crate handling.

### Part Three

Now, we wish to extend the simulation to allow representation a new kind of robot which is able to travel diagnally when traversing the warehouse grid:

The supported command syntax for the simulated robot should remain the same, but if the robot is issued a pair of commands which would result in it moving (for example) North and then East, it should instead simply perform a single North-East movement.

Provide tests to validate that the new simulated robot performs correctly.
