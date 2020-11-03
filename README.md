# Robot Warehouse

## Scenario

We are installing a new type of robot into our (hypothetical) warehouse as part of an automation project.  As part of this project, there are various software components which need to be developed.

## About the Robots

For convenience the robot moves along a grid in the roof of the warehouse and we have made sure that all of our warehouses are built so that the dimensions of the grid are 10 by 10; objects in the warehouse, including the robots, are always aligned with the grid, so object locations' may be treated as integer coordinates.  We've also made sure that all our warehouses are aligned along north-south and east-west axes. The system operates on a cartesian coordinate map that aligns to the warehouse's physical dimensions: point (0, 0) indicates the most south-west and (10, 10) indicates the most north-east.

Each robot operates by being given 'tasks' which each consist of a string of 'commands':

All of the commands to the robot consist of a single capital letter and different commands are optionally delineated by whitespace.

The robot should accept the following commands:

- N move one unit north
- W move one unit west
- E move one unit east
- S move one unit south

Example command sequences:

* The command sequence: `"N E S W"` will move the robot in a full square, returning it to where it started.

* If the robot starts in the south-west corner of the warehouse then the following commands will move it to the middle of the warehouse: `"N E N E N E N E"`

The robot will only perform a single task at a time: if additional tasks are given to the robot while is busy performing a task, those additional tasks are queued up, and will be executed once the preceding task is completed (or aborted for some reason).  Each task is identified with a unique string ID, and a task which is either in progress or enqueued can be aborted/cancelled at any time.  If the robot is unable to execute a particular command (for instance, because the command would cause the robot to run into the edges of the warehouse grid) then an error occurs, and the entire task is aborted.
