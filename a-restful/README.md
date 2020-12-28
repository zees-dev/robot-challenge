# Robot Warehouse

We have installed a robot in our warehouse and now we need to send it commands to control it. We need you to implement the high level RESTful APIs, which can be called from a ground control station.

For convenience the robot moves along a grid in the roof of the warehouse and we have made sure that all of our warehouses are built so that the dimensions of the grid are 10 by 10. We've also made sure that all our warehouses are aligned along north-south and east-west axes. The robot also builds an internal x y coordinate map that aligns to the warehouse's physical dimensions. On the map, point (0, 0) indicates the most south-west and (10, 10) indicates the most north-east.

All of the commands to the robot consist of a single capital letter and different commands are delineated by whitespace.

The robot should accept the following commands:

- N move north
- W move west
- E move east
- S move south

Example command sequences
The command sequence: "N E S W" will move the robot in a full square, returning it to where it started.

If the robot starts in the south-west corner of the warehouse then the following commands will move it to the middle of the warehouse.

"N E N E N E N E"

## Robot SDK Interface

The robot provids a set of low level SDK functions in GO to control its movement.

```go
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

- Create a RESTful API to accept a series of commands to the robot.
- Make sure that the robot doesn't try to move outside the warehouse.
- Create a RESTful API to report the command series's execution status.
- Create a RESTful API cancel the command series.
- The RESTful service should be written in Golang.

## Challenge

- The Robot SDK is still under development, you need to find a way to prove your API logic is working.
- The ground control station wants to be notified as soon as the command sequence completed. Please provide a high level design overview how you can achieve it. This overview is not expected to be hugely detailed but should clearly articulate the fundamental concept in your design.

---

## Implementation

...

### Assumptions

- There is no time taken to execute a sequence of commands (assuming they are valid and  can be performed)
  - Hence it is probably only possible to cancel an in-flight command if the server is receiving too many commands and the desired command associated to taskID has not been executed yet (still in channel queue)
- There is only a single robot operating on the roof (registrations and/or collisions with  other robots is out of scope)

### Features

- This solution should enable high-throughput, concurrent safe robot operations
- OpenAPI spec for Restful API calls
- Pluggable storage (in-memory map, database or any other storage) - achieved via implementation of repository interface
- Serverside event support - a client can subscribe to HTTP2 SSE to get real-time updates of robot state

## 3rd party modules

- [uuid](github.com/satori/go.uuid) - for unique taskID generation
- [gorilla mux](github.com/gorilla/mux) - http request multiplexer (standard library compliant)

## TODO

- [ ] Implement Restful API endpoints
  - Move robot (PUT)
    - /move
    - 200 (ok) - taskId & success/failure, 400 (bad request)
    - Note: Use context
  - Get list of commands sent to robot (taskId) with status (success/failed) (GET)
    - /tasks
    - 200 (ok)
  - Get single command status by taskId (GET)
    - /task/{id}
    - 200 (ok), 404 (command sequence with taskId not found)
  - Cancel command series (Delete)
    - /task/{id}
    - 204 (no content), 404 (command sequence with taskId not found)

- [ ] Challenge SSE (server-sent-events) which sends -> taskId, status, robot final state
  - Write up design/architecture doc (SSE, HTTP/2, browser support)

- [ ] OpenAPI compliant spec
  - Have a look at Twirp, GRPC-gateway
  - Have a look at Go Kit - the transport should a decoupled part of the architecture

- [ ] Testing
  - [ ] Implement unit tests for functionality
  - [ ] Implement integrations tests for API
  - [ ] Implement test coverage

- [ ] Challenge
  - Implement minimal frontend or console UI to view state of a warehouse
    - Serve this static SPA or directory from backend API
  - A `writer` should get notified of successful command completion (potentially write output to file)
    - Write up design/architecture doc (SSE, HTTP/2, browser support)
    - The writer sends server-side event to an admin panel?
      - SSE (server-sent-events) which sends -> taskId, status, robot final state

- [ ] Dockerize API

## Improvements

- [ ] Implement Auth (JWT using bearer scheme)
- [ ] Persist robot operations to a database (sqlite will do)
- [ ] Distribute to [pkg.go.dev](https://pkg.go.dev/) for open source projects

## Self-defined constraints & assumptions

- Robos should not be able to collide with each other (error thrown if/when this happens)
- The robot takes a certain time to move - this has been defaulted to 1 second per move
- A 10x10 grid south-west = (0,0), while north-east = (9,9) - instead of (10,10)
- `"N E N E N E N E"` ends up at (4,4) - there cannot be a true `center` for a 10x10 grid

## Architecture

- 2 servers
  - RESTful Admin API server
    - Deals with warehouse/bot creation
    - Responsible for spinning up warehouse server(s)
    - This server can double up as Auth server
  - RESTful Warehouse API server
    - Responsible for bot control within a warehouse
    - Service contains in-memory state of a warehouse
    - Successful bot state changes written to DB
    - Ideally this server should use an IoT optimized protocol - gRPC would be perfect here

## Rest Endpoints

Move bot:

```sh
curl \
  -d '{"commands": "N E N E"}' \
  -X PUT localhost:8000/move
```

<!-- Get all warehouses

```sh
curl localhost:8000/warehouse
```

Create warehouse:

```sh
curl \
  -d '{"id": 1, "width": 10, "height": 10}' \
  -X POST localhost:8000/warehouse
```

Get warehouse by id

```sh
curl localhost:8000/warehouse/1
```

Create bot:

```sh
curl \
  -d '{"id": 1, "warehouseId": 1, "x": 1, "y": 1}' \
  -X POST localhost:8000/bot
```

```sh
curl \
  -d '{"id": 1, "warehouseId": 1, "x": 1, "y": 1}' \
  -X POST localhost:8000/bot
``` -->
