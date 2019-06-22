package main

import (
	"bytes"
	"encoding/json"
	"github.com/jinzhu/configor"
	"github.com/satori/go.uuid"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

var a App

func TestMain(m *testing.M) {
	configor.Load(&config, "config-test.yml")
	a = App{}
	a.Initialize(config)
	taskRunner = newTaskRunner(config, a.DB)

	code := m.Run()

	clearCollection()

	os.Exit(code)
}

func TestWrongGUID(t *testing.T) {
	clearCollection()

	req, _ := http.NewRequest("GET", "/task/123", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusBadRequest, response.Code)
}

func TestGetNonExistentTask(t *testing.T) {
	clearCollection()

	req, _ := http.NewRequest("GET", "/task/8ada3cf8-c782-4866-9d44-0c5b39b39487", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)
}

func TestGetRunningTask(t *testing.T) {
	clearCollection()

	createTaskResp := createTask()

	time.Sleep(time.Duration(config.TaskExecutionTimeSeconds-1) * time.Second)

	req, _ := http.NewRequest("GET", "/task/"+createTaskResp["guid"], nil)
	response := executeRequest(req)

	var getTaskResp map[string]string
	json.Unmarshal(response.Body.Bytes(), &getTaskResp)

	if getTaskResp["status"] != string(Running) {
		t.Errorf("Expected status = running. Got '%v'", getTaskResp["status"])
		return
	}
}

func TestGetFinishedTask(t *testing.T) {
	clearCollection()

	createTaskResp := createTask()

	time.Sleep(time.Duration(config.TaskExecutionTimeSeconds+1) * time.Second)

	req, _ := http.NewRequest("GET", "/task/"+createTaskResp["guid"], nil)
	response := executeRequest(req)

	var getTaskResp map[string]string
	json.Unmarshal(response.Body.Bytes(), &getTaskResp)

	if getTaskResp["status"] != string(Finished) {
		t.Errorf("Expected status = finished. Got '%v'", getTaskResp["status"])
		return
	}
}

func createTask() map[string]string {
	payload := []byte(``)
	req, _ := http.NewRequest("POST", "/task", bytes.NewBuffer(payload))
	response := executeRequest(req)
	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	return m
}

func TestCreateTask(t *testing.T) {
	clearCollection()

	payload := []byte(``)

	req, _ := http.NewRequest("POST", "/task", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusAccepted, response.Code)

	var createTaskResp map[string]string
	json.Unmarshal(response.Body.Bytes(), &createTaskResp)

	_, err := uuid.FromString(createTaskResp["guid"])
	if err != nil {
		t.Errorf("Expected to get valid guid. Got '%v'", createTaskResp["guid"])
		return
	}
}

func clearCollection() {
	a.DB.Session.DB(config.MongoDatabase).C(config.Collection).RemoveAll(nil)
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}
