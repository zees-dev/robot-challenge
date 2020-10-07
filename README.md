# Robot Warehouse

We have installed a robot in our warehouse and now we need to be able to send it commands to control it. We need you to implement the high level restful apis, which can be called from ground control station.

For convenience the robot moves along a grid in the roof of the warehouse and we have made sure that all of our warehouses are built so that the dimensions of the grid are 10 by 10. We've also made sure that all our warehouses are aligned along north-south and east-west axes. The robot also build an internal x y coordinates map that aligns the warehouse. On the map, point (0, 0) indicates the most south-west and (10, 10) indicates the most north-east.

All of the commands to the robot consist of a single capital letter and different commands are dilineated by whitespace.

The robot should accept the following commands:

- N move north
- W move west
- E move east
- S move south

Example command sequences
The command sequence: "N E S W" will move the robot in a full square, returning it to where it started.

If the robot starts in the south-west corner of the warehouse then the following commands will move it to the middle of the warehouse.

"N E N E N E N E"

## Robot SDK Commands 

The robot provids a set of low level SDK functions in GO to control its movement. 

- `func MoveTo(x, y float) (taskID string, taskComplete chan error)` Requests robot move to an absolute position x, y on the map. 
    - `taskID`: Unique task identifier 
    - `taskComplete`: The event indicates the task was completed 
- `func Cancel(taskID string) error` Requests robot to cancel the task.
- `func CurrentPosition() (x, y float)` Returns the absolute position value on the map.

## Requirements
- Create a restful api to accept a series of commands to the robot. 
- Make sure that the robot doesn't try to move outside the warehouse.
- Create a restful api to report commands series exection status.
- Create a restful api cancel the commands series.

## Challenge
- The Robot SDK is still under development, you need to find a way to prove your api logic is working.
- The ground control station wants to be notified whenever the command sequence completed. Please provide a high level design overview how you can achieve it.
