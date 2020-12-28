package main

import (
	"flag"
	"log"
)

func main() {
	xPtr := flag.Uint("x", 0, "robot initialisation x co-ordinate")
	yPtr := flag.Uint("y", 0, "robot initialisation y co-ordinate")
	xDimension, yDimension := uint(10), uint(10)
	flag.Parse()

	x, y := *xPtr, *yPtr
	if x >= xDimension {
		log.Fatalf("Invalid robot x position; x co-ordinate must satisfy 0 <= x < %d", xDimension)
	}
	if y >= yDimension {
		log.Fatalf("Invalid robot x position; y co-ordinate must satisfy 0 <= y < %d", yDimension)
	}
	// TODO (improvement) - check if not being registered on an existing robos location (if multi-robots are supported)

	// db := NewDB()
	// latestState, err := db.GetLatestState()
	// if err != nil {
	// 	log.Println(err)
	// } else {
	// 	x, y = latestState.X, latestState.Y
	// }

	robot := NewBot(x, y)
	go robot.RunRobot()
	log.Printf("Initialising robot at (%d, %d)...", x, y)
	// robotSvc := NewRobotService(db, robot)

	// TODO run static file server
	// TODO serve OpenAPI spec from static file server
	// TODO serve minimal frontend from static file server

	robotAPIServer(robot)

	// TODO look into graceful server shutdown (OS signals)
}
