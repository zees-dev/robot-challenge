package main

import "log"

func main() {
	// TODO - initialise robot in x,y within 10 - using flags
	// TODO - check if within bounds
	// TODO (improvement) - check if not being registered on an existing robos location (if multi-robots are supported)
	x, y := uint(0), uint(0)

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

	robotServer(robot)

	// TODO look into graceful server shutdown (OS signals)
}
