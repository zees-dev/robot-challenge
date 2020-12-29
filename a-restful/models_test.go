package main

import (
	"errors"
	"sync"
	"testing"
)

func TestValidateCommandSequence(t *testing.T) {
	t.Run("test invalid character", func(t *testing.T) {
		commands := "A N"
		err := validateCommandSequence(commands)
		if err == nil {
			t.Errorf("command `A` is invalid")
		}
	})

	t.Run("test invalid empty string", func(t *testing.T) {
		commands := ""
		err := validateCommandSequence(commands)
		if err == nil {
			t.Errorf("empty string input is invalid")
		}
	})

	t.Run("test invalid whitespace string", func(t *testing.T) {
		commands := " "
		err := validateCommandSequence(commands)
		if err == nil {
			t.Errorf("whitespace string input is invalid")
		}
	})

	t.Run("test valid input commands", func(t *testing.T) {
		commands := "N S E W"
		err := validateCommandSequence(commands)
		if err != nil {
			t.Errorf("command sequence is valid")
		}
	})
}

func TestGetUpdatedState(t *testing.T) {
	bot := NewBot(0, 0, NewInMemoryDB())

	t.Run("test `S` command seq failure at (0,0)", func(t *testing.T) {
		commands := "S"
		_, err := bot.getUpdatedState(commands)
		if err == nil {
			t.Errorf("command `S` should not be performed")
		}
	})

	t.Run("test `W` command seq failure at (0,0)", func(t *testing.T) {
		commands := "W"
		_, err := bot.getUpdatedState(commands)
		if err == nil {
			t.Errorf("command `W` should not be performed")
		}
	})

	t.Run("test `S N` command seq fails if even one command exceeds warehouse dimensions while final output does not", func(t *testing.T) {
		commands := "S N"
		_, err := bot.getUpdatedState(commands)
		if err == nil {
			t.Errorf("command `S N` should not be performed")
		}
	})

	t.Run("test `N` command seq success at (0,0) - moves robot  to (0,1)", func(t *testing.T) {
		commands := "N"
		got, err := bot.getUpdatedState(commands)
		if err != nil {
			t.Errorf("command `N` should be performed")
		}

		want := RobotState{0, 1, false}
		if got != want {
			t.Errorf("command `N` should set robot at (0,1)")
		}
	})

	t.Run("test `E` command seq success at (0,0) - moves robot  to (1,0)", func(t *testing.T) {
		commands := "E"
		got, err := bot.getUpdatedState(commands)
		if err != nil {
			t.Errorf("command `E` should be performed")
		}

		want := RobotState{1, 0, false}
		if got != want {
			t.Errorf("command `E` should set robot at (1,0)")
		}
	})

	t.Run("test `N E S W` command seq success at (0,0) - returns robot  to original position", func(t *testing.T) {
		commands := "N E S W"
		got, err := bot.getUpdatedState(commands)
		if err != nil {
			t.Errorf("commands `N E S W` should be performed")
		}

		want := RobotState{0, 0, false}
		if got != want {
			t.Errorf("commands `N E S W` should set robot back at (0,0)")
		}
	})

	t.Run("test `N E N E N E N E` command seq success at (0,0) - moves robot to (4,4)", func(t *testing.T) {
		commands := "N E N E N E N E"
		got, err := bot.getUpdatedState(commands)
		if err != nil {
			t.Errorf("commands `N E N E N E N E` should be performed")
		}

		want := RobotState{4, 4, false}
		if got != want {
			t.Errorf("commands `N E N E N E N E` should move robot to (4,4)")
		}
	})
}

func TestUpdateCurrentState(t *testing.T) {
	bot := NewBot(0, 0, NewInMemoryDB())

	t.Run("test (10,0) is invalid robot state", func(t *testing.T) {
		rs := RobotState{10, 0, false}
		err := bot.UpdateCurrentState(rs)
		if err == nil {
			t.Errorf("incoming robot state (10,0) should not be set")
		}
	})

	t.Run("test (0,10) is invalid robot state", func(t *testing.T) {
		rs := RobotState{0, 10, false}
		err := bot.UpdateCurrentState(rs)
		if err == nil {
			t.Errorf("incoming robot state (0,10) should not be set")
		}
	})

	t.Run("test (9,0) is valid robot state", func(t *testing.T) {
		rs := RobotState{9, 0, false}
		err := bot.UpdateCurrentState(rs)
		if err != nil {
			t.Errorf("incoming robot state (9,0) should be set")
		}
	})

	t.Run("test (0,9) is valid robot state", func(t *testing.T) {
		rs := RobotState{0, 9, false}
		err := bot.UpdateCurrentState(rs)
		if err != nil {
			t.Errorf("incoming robot state (0,9) should be set")
		}
	})

	t.Run("test (9,9) is valid robot state", func(t *testing.T) {
		rs := RobotState{9, 9, false}
		err := bot.UpdateCurrentState(rs)
		if err != nil {
			t.Errorf("incoming robot state (9,9) should be set")
		}
	})
}

func TestCurrentState(t *testing.T) {
	bot := NewBot(0, 0, NewInMemoryDB())

	t.Run("test (9,9) successfully updates robot state", func(t *testing.T) {
		rs := RobotState{9, 9, false}
		bot.UpdateCurrentState(rs)

		updatedState := bot.CurrentState()
		if rs != updatedState {
			t.Errorf("robot should successfully move to (9,9)")
		}
	})
}

func TestEnqueueTask(t *testing.T) {
	bot := NewBot(0, 0, NewInMemoryDB())

	t.Run("test successfully generates taskID", func(t *testing.T) {
		go func() { <-bot.tasks }()
		taskID, _, _ := bot.EnqueueTask("N S E W")
		if taskID == "" {
			t.Error("robot should have a queued task")
		}
	})

	t.Run("test successfully queues taskID", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(1)

		var got string
		go func() { got = <-bot.tasks; wg.Done() }()

		want, _, _ := bot.EnqueueTask("N S E W")

		wg.Wait()

		if want != got {
			t.Errorf("robot should have a queued task; got: \"%s\", want \"%s\"", got, want)
		}
	})
}

func TestCancelTask(t *testing.T) {
	t.Run("test successfully cancels non-executed task", func(t *testing.T) {
		bot := NewBot(0, 0, NewInMemoryDB())
		go func() { <-bot.tasks }()

		commandSeq := "N E S W"
		taskID, _, _ := bot.EnqueueTask(commandSeq)

		err := bot.CancelTask(taskID)

		if err != nil {
			t.Errorf("should successfully execute non-executed task %s", taskID)
		}
	})

	t.Run("test fails to find non-existent task", func(t *testing.T) {
		bot := NewBot(0, 0, NewInMemoryDB())
		go bot.RunRobot()

		commandSeq := "N E S W"
		bot.EnqueueTask(commandSeq)

		err := bot.CancelTask("cdc29b67-7212-4579-a593-74fb9a1f606f")
		if err == nil {
			t.Error("task ID cdc29b67-7212-4579-a593-74fb9a1f606f should not be found")
		}
	})

	t.Run("test failed to cancel pre-executed task", func(t *testing.T) {
		bot := NewBot(0, 0, NewInMemoryDB())
		go bot.RunRobot()

		commandSeq := "N E S W"
		taskID, _, _ := bot.EnqueueTask(commandSeq)

		rTask, _ := bot.repository.GetTask(taskID)
		rTask.executed = true
		bot.repository.UpdateTask(rTask)

		err := bot.CancelTask(taskID)
		if err == nil {
			t.Errorf("should fail to cancel executed task %s", taskID)
		}
	})
}

// TestRobotMovementSubscriptions provides an insight of how consumers of the `position` channel can subscribe to robot state changes
func TestRobotMovementSubscriptions(t *testing.T) {
	bot := NewBot(0, 0, NewInMemoryDB())
	go bot.RunRobot()

	var wg sync.WaitGroup
	wg.Add(1)

	_, position, _ := bot.EnqueueTask("N E N E N E N E")
	var got RobotState
	go func() { got = <-position; wg.Done() }()

	wg.Wait()

	want := RobotState{4, 4, false}

	if got != want {
		t.Errorf("robot should have updated state; got: %v, want: %v", got, want)
	}
}

// TestRobotErrorSubscriptions provides an insight of how consumers of the `err` channel can subscribe to invalid robot state changes
func TestRobotErrorSubscriptions(t *testing.T) {
	bot := NewBot(0, 0, NewInMemoryDB())
	go bot.RunRobot()

	var wg sync.WaitGroup
	wg.Add(1)

	_, _, err := bot.EnqueueTask("N E S S")
	var got error
	go func() { got = <-err; wg.Done() }()

	wg.Wait()

	want := errors.New("command `S` exceeds warehouse dimensions")

	if got.Error() != want.Error() {
		t.Errorf("robot movement should have thrown error; got: \"%s\", want: \"%s\"", got, want)
	}
}
