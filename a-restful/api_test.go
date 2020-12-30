package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func getHTTPHandler() http.Handler {
	robot := NewBot(0, 0, NewInMemoryDB())
	go robot.listen()
	handler := RobotAPIServer(&robot)
	return handler
}

func TestHealthEndpoint(t *testing.T) {
	handler := getHTTPHandler()
	rr := httptest.NewRecorder()

	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	handler.ServeHTTP(rr, req)

	// Check response status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	want := `{"status":"healthy"}`
	got := rr.Body.String()
	if got != want {
		t.Errorf("incorrect health response; got: %s, want: %s", got, want)
	}
}

func TestGetRobotStateEndpointSuccess(t *testing.T) {
	handler := getHTTPHandler()
	rr := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/api/v1/state", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	handler.ServeHTTP(rr, req)

	// check response status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// unmarshal response body
	var responseBody map[string]uint
	json.Unmarshal(rr.Body.Bytes(), &responseBody)

	fmt.Println(responseBody)

	// check if `taskID` exists in json response
	x, ok := responseBody["x"]
	if !ok {
		t.Error("response must contain `x`")
	}
	if x != 0 {
		t.Error("x co-ordinate should be 0")
	}

	y, ok := responseBody["y"]
	if !ok {
		t.Error("response must contain `y`")
	}
	if y != 0 {
		t.Error("y co-ordinate should be 0")
	}
}

func TestMoveRobotEndpointSuccess(t *testing.T) {
	handler := getHTTPHandler()
	rr := httptest.NewRecorder()
	req, err := http.NewRequest("PUT", "/api/v1/state", bytes.NewBuffer([]byte(`{"commands":"N E N E"}`)))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	handler.ServeHTTP(rr, req)

	// check response status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// unmarshal response body
	var responseBody map[string]string
	json.Unmarshal(rr.Body.Bytes(), &responseBody)

	// check if `taskID` exists in json response
	if _, ok := responseBody["taskID"]; !ok {
		t.Error("response must contain `taskID`")
	}
}

func TestMoveRobotEndpointInvalidCommands(t *testing.T) {
	handler := getHTTPHandler()
	rr := httptest.NewRecorder()

	req, err := http.NewRequest("PUT", "/api/v1/state", bytes.NewBuffer([]byte(`{"commands":"N E N A"}`)))
	if err != nil {
		t.Error(err)
	}

	handler.ServeHTTP(rr, req)

	// check response status code
	statusWant := http.StatusBadRequest
	statusGot := rr.Code
	if statusWant != statusGot {
		t.Errorf("handler returned wrong status code; want %v, got %v", statusWant, statusGot)
	}

	// check content type
	contentTypeWant := "text/plain; charset=utf-8"
	contentTypeGot := rr.Result().Header.Get("Content-Type")
	if contentTypeWant != contentTypeGot {
		t.Errorf(`incorrect content type header response; want: "%s", got: "%s"`, contentTypeWant, contentTypeGot)
	}

	// check response
	resWant := `invalid command 'A', command can only be one of 'N', 'S', 'E' or 'W'`
	resGot := strings.TrimSpace(rr.Body.String())
	if resWant != resGot {
		t.Errorf(`incorrect error response; want: "%s", got: "%s"`, resWant, resGot)
	}
}

func TestMoveRobotEndpointEmptyCommands(t *testing.T) {
	handler := getHTTPHandler()
	rr := httptest.NewRecorder()

	req, err := http.NewRequest("PUT", "/api/v1/state", bytes.NewBuffer([]byte(`{"commands":" "}`)))
	if err != nil {
		t.Error(err)
	}

	handler.ServeHTTP(rr, req)

	// check response status code
	statusWant := http.StatusBadRequest
	statusGot := rr.Code
	if statusWant != statusGot {
		t.Errorf("handler returned wrong status code; want %v, got %v", statusWant, statusGot)
	}

	// check content type
	contentTypeWant := "text/plain; charset=utf-8"
	contentTypeGot := rr.Result().Header.Get("Content-Type")
	if contentTypeWant != contentTypeGot {
		t.Errorf(`incorrect content type header response; want: "%s", got: "%s"`, contentTypeWant, contentTypeGot)
	}

	// check response
	resWant := `failed to execute empty commands`
	resGot := strings.TrimSpace(rr.Body.String())
	if resWant != resGot {
		t.Errorf(`incorrect error response; want: "%s", got: "%s"`, resWant, resGot)
	}
}

func TestMoveRobotEndpointMultipleWhitespaceCommands(t *testing.T) {
	handler := getHTTPHandler()
	rr := httptest.NewRecorder()

	req, err := http.NewRequest("PUT", "/api/v1/state", bytes.NewBuffer([]byte(`{"commands":"N  E N"}`)))
	if err != nil {
		t.Error(err)
	}

	handler.ServeHTTP(rr, req)

	// check response status code
	statusWant := http.StatusBadRequest
	statusGot := rr.Code
	if statusWant != statusGot {
		t.Errorf("handler returned wrong status code; want %v, got %v", statusWant, statusGot)
	}

	// check content type
	contentTypeWant := "text/plain; charset=utf-8"
	contentTypeGot := rr.Result().Header.Get("Content-Type")
	if contentTypeWant != contentTypeGot {
		t.Errorf(`incorrect content type header response; want: "%s", got: "%s"`, contentTypeWant, contentTypeGot)
	}

	// check response
	resWant := `invalid command 'N  E N'; command  cannot contain multiple whitespaces`
	resGot := strings.TrimSpace(rr.Body.String())
	if resWant != resGot {
		t.Errorf(`incorrect error response; want: "%s", got: "%s"`, resWant, resGot)
	}
}

func TestGetTaskEndpointSuccess(t *testing.T) {
	handler := getHTTPHandler()
	rr := httptest.NewRecorder()

	taskID := func() string {
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/api/v1/state", bytes.NewBuffer([]byte(`{"commands":"N E"}`)))
		handler.ServeHTTP(rr, req)
		var responseBody map[string]string
		json.Unmarshal(rr.Body.Bytes(), &responseBody)
		return responseBody["taskID"]
	}()

	req, err := http.NewRequest("GET", fmt.Sprintf("/api/v1/task/%s", taskID), nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(rr, req)

	// check response status code
	statusWant := http.StatusOK
	statusGot := rr.Code
	if statusWant != statusGot {
		t.Errorf("handler returned wrong status code; want %v, got %v", statusWant, statusGot)
	}

	// unmarshal response body
	var responseBody map[string]map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &responseBody)

	// check response
	if _, ok := responseBody["task"]; !ok {
		t.Error("response must contain `task`")
	}
	if _, ok := responseBody["task"]["id"]; !ok {
		t.Error("response must contain `task`->`id`")
	}
	if _, ok := responseBody["task"]["command"]; !ok {
		t.Error("response must contain `task`->`command`")
	}
	if _, ok := responseBody["task"]["executed"]; !ok {
		t.Error("response must contain `task`->`executed`")
	}
	if _, ok := responseBody["task"]["cancelled"]; !ok {
		t.Error("response must contain `task`->`cancelled`")
	}
}

func TestGetTaskEndpointNotFound(t *testing.T) {
	handler := getHTTPHandler()
	rr := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/api/v1/task/non-existent", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	handler.ServeHTTP(rr, req)

	// check response status code
	statusWant := http.StatusNotFound
	statusGot := rr.Code
	if statusWant != statusGot {
		t.Errorf("handler returned wrong status code; want %v, got %v", statusWant, statusGot)
	}

	// check response
	resWant := `Task with ID 'non-existent' not found`
	resGot := strings.TrimSpace(rr.Body.String())
	if resWant != resGot {
		t.Errorf(`incorrect error response; want: "%s", got: "%s"`, resWant, resGot)
	}
}

// TestDeleteTaskEndpointSuccess is tested without `getHTTPHandler` since we need to manually instantiate server without `robot.listen
// This is because order of operations for cancel task is not guaranteed if `robot.listen` is running (as goroutine can process/modify task)
// which can prevent task cancellation (since task may already be executed)
func TestDeleteTaskEndpointSuccess(t *testing.T) {
	robot := NewBot(0, 0, NewInMemoryDB())
	go func() { <-robot.tasks }() // prevent channel blocking
	handler := RobotAPIServer(&robot)

	rr := httptest.NewRecorder()

	taskID := func() string {
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/api/v1/state", bytes.NewBuffer([]byte(`{"commands":"N E"}`)))
		handler.ServeHTTP(rr, req)
		var responseBody map[string]string
		json.Unmarshal(rr.Body.Bytes(), &responseBody)
		return responseBody["taskID"]
	}()

	req, err := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/task/%s", taskID), nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	handler.ServeHTTP(rr, req)

	// check response status code
	statusWant := http.StatusNoContent
	statusGot := rr.Code
	if statusWant != statusGot {
		t.Errorf("handler returned wrong status code; want %v, got %v", statusWant, statusGot)
	}
}

func TestDeleteTaskEndpointNotFound(t *testing.T) {
	handler := getHTTPHandler()
	rr := httptest.NewRecorder()
	req, err := http.NewRequest("DELETE", "/api/v1/task/non-existent", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	handler.ServeHTTP(rr, req)

	// check response status code
	statusWant := http.StatusNotFound
	statusGot := rr.Code
	if statusWant != statusGot {
		t.Errorf("handler returned wrong status code; want %v, got %v", statusWant, statusGot)
	}

	// check response
	resWant := `Task with ID 'non-existent' not found`
	resGot := strings.TrimSpace(rr.Body.String())
	if resWant != resGot {
		t.Errorf(`incorrect error response; want: "%s", got: "%s"`, resWant, resGot)
	}
}
