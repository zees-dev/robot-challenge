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

## Solution

The solution can be run from the __a-restful__ directory.\
The server is run at port _8000_.

**Install go modules:**

```sh
go mod download
```

**Run Restful robot server:**

```sh
go run .
```

## Run via Docker

**Build:**

```sh
docker build -t rocos-robot:alpine .
```

**Run:**

```sh
docker run --rm -it -p 8000:8000 rocos-robot:alpine
```

### Command line flags

The `x` and `y` flags can optionally be passed in upon running the robot server to set the initial robot position on the warehouse - these currectly default to `0, 0` respectively.

**Example - initialising robot at position (5, 2):**

```sh
go run . -x 5 -y 2
```

### Frontend

A minimal browser based frontend/client is served at [http://localhost:8000/](http://localhost:8000/) which allows one to visually interact with the robot server APIs.

- This has been included the demonstrate API functionality
- The frontend also implements a POC of the challenge (realtime robot notifications)  to demonstrate high level design viability of robot notifications.

**Note:** This is an implementation of the challenge

### Open API

The robot server is Open API compliant; and hence serves  the Open API spec to enable API interactivity and testing from the browser. The spec is served at [http://localhost:8000/swaggerui/](http://localhost:8000/swaggerui/).

**Note:** Only the [spec](./swaggerui/swagger.json) was designed from the [swagger editor](https://editor.swagger.io/); the UI files to render the spec have been obtained from [official swagger repo](https://github.com/swagger-api/swagger-ui/tree/master/dist).

## Implementation assumptions

- There is no time taken to execute a sequence of commands (assuming they are valid and  can be performed)
  - Hence it is probably only possible to cancel an in-flight command if the server is receiving too many commands and the desired command associated to taskID has not been executed yet (still in channel queue)
- There is only a single robot operating on the roof (registrations and/or collisions with  other robots is out of scope)

**Note:** The API does not consume the `Robot` SDK interface since a get task by ID method is required to fulfil requirements; the `Robot` interface does not have such a method...

## Features

- This solution should enable high-throughput, concurrent safe robot operations
- Interactive frontend to view robot state on a grid (in browser) - in realtime
  - 
- OpenAPI spec for Restful API calls
- Pluggable storage (in-memory map, database or any other storage) - achieved via implementation of repository interface
- Server-Sent Events support - a client can subscribe to SSE to get real-time updates of robot state
  - This is a POC of proposed solution to real-time notifications (challenge)
  - The provided frontend is using this solution to get notified/updated in real-time (upon robot state changes)

## Testing

[Unit tests](./models_test.go) and [integration tests](./api_test.go) have been implemented and can be run using:

```sh
go test .
```

### Test race conditions

```sh
go test -race .
```

Note: The command above requires `CGO_ENABLED=1`

### Test with coverage

```sh
go test . -coverprofile cp.out
cat cp.out | grep -v "storage.go" > cover.out
go tool cover -func cover.out
```

**Note:** The `storage.go` file is ignored from coverage since the storage is de-coupled/pluggable (assuming the `Repository` interface is implemented)

## 3rd party modules

- [uuid](github.com/satori/go.uuid) - for unique taskID generation
- [gorilla mux](github.com/gorilla/mux) - http request multiplexer (standard library compliant)

## Improvements

- Implement Auth (JWT using bearer scheme)
- Migrate to gRPC since its lower latency & bandwidth - hence best suited for thid usecase
- Use Go 1.6 (currently in beta; released in Q1 2021) to embed static files into single binary (new feature)
- Persist robot operations to a database (sqlite will do) - implement persistent repository
- Distribute to [pkg.go.dev](https://pkg.go.dev/) for open source projects

## Rest Endpoints

### Health

```sh
curl -X GET 'http://localhost:8000/health'
```

### Get bot state

```sh
curl -X GET 'http://localhost:8000/api/v1/state'
```

### Update bot state

```sh
curl \
  -d '{"commands": "N E N E"}' \
  -X PUT 'http://localhost:8000/api/v1/state'
```

### Get command execution status

```sh
curl -X GET 'http://localhost:8000/api/v1/task/<task-id>'
```

### Cancel queue command sequence

```sh
curl -X DELETE 'http://localhost:8000/api/v1/task/<task-id>'
```

### Subscribe to real-time robot state updates

```sh
curl -X GET 'http://localhost:8000/api/v1/state/subscribe'
```

---

## TODO

- [x] Implement Restful API endpoints
  - Move robot (PUT)
    - /state
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
- [ ] Implement context based request cancellation

- [x] OpenAPI compliant spec
  - Serve openapi file using static server

- [x] Testing
  - [x] Implement unit tests for functionality
  - [x] Implement integrations tests for API
  - [x] Implement test coverage
  - [x] Check for race conditions

- [x] Challenge
  - Implement minimal frontend or console UI to view state of a warehouse
    - Serve this static SPA or directory from backend API
  - Technical design document proposing how clients can gett notified to robot state changes

- [x] Implement challenge POC

- [x] Dockerize API
