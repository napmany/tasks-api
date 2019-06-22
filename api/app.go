package main

import (
	"encoding/json"
	"github.com/globalsign/mgo"
	"github.com/gorilla/mux"
	"github.com/satori/go.uuid"
	"log"
	"net/http"
	"time"
)

const (
	InternalErrorMessage = "Internal Server Error"
	WrongGUIDMessage     = "It's not a GUID"
	NotFoundMessage      = "Not found"
)

var taskRunner *TaskRunner

type App struct {
	Router *mux.Router
	DB     *DB
}

type DB struct {
	Session *mgo.Session
}

func (a *App) Initialize(cfg *Config) {
	a.DB = initialiseMongo(config)
	taskRunner = newTaskRunner(config, a.DB)
	a.Router = mux.NewRouter().StrictSlash(true)
	a.initializeRoutes()
}

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/task/{guid}", a.taskGetHandler).Methods("GET")
	a.Router.HandleFunc("/task", a.taskPostHandler).Methods("POST")
}

func (a *App) Run(cfg *Config) {
	log.Fatal(http.ListenAndServe(config.Addr, a.Router))
}

func initialiseMongo(cfg *Config) (db *DB) {
	info := &mgo.DialInfo{
		Addrs:    []string{cfg.MongoHosts},
		Timeout:  60 * time.Second,
		Database: cfg.MongoDatabase,
		Username: cfg.MongoUsername,
		Password: cfg.MongoPassword,
	}

	session, err := mgo.DialWithInfo(info)
	if err != nil {
		panic(err)
	}

	db = new(DB)
	db.Session = session
	return
}

func (a *App) taskGetHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	guid := params["guid"]
	id, err := uuid.FromString(guid)
	if err != nil {
		log.Println(err)
		http.Error(w, WrongGUIDMessage, http.StatusBadRequest)
		return
	}

	task, err := findTaskByGUID(a.DB, id)

	if err == mgo.ErrNotFound {
		log.Println(err)
		http.Error(w, NotFoundMessage, http.StatusNotFound)
		return
	} else if err != nil {
		log.Println(err)
		http.Error(w, InternalErrorMessage, http.StatusInternalServerError)
		return
	}

	response := make(map[string]string)
	response["status"] = string(task.Status)
	response["timestamp"] = task.Timestamp.Format(time.RFC3339)
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		log.Println(err)
		http.Error(w, InternalErrorMessage, http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func (a *App) taskPostHandler(w http.ResponseWriter, r *http.Request) {
	col := a.DB.Session.DB(config.MongoDatabase).C(config.Collection)
	task := newTask()
	err := col.Insert(task)
	if err != nil {
		log.Println(err)
		http.Error(w, InternalErrorMessage, http.StatusInternalServerError)
		return
	}

	go taskRunner.run(task.GUID.UUID)

	response := make(map[string]string)
	response["guid"] = task.GUID.String()
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		log.Println(err)
		http.Error(w, InternalErrorMessage, http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write(jsonResponse)
}
