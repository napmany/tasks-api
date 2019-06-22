package main

import (
	"github.com/globalsign/mgo/bson"
	"github.com/satori/go.uuid"
	"log"
	"time"
)

type Status string

const (
	Created  Status = "created"
	Running  Status = "running"
	Finished Status = "finished"
)

type Task struct {
	ID        bson.ObjectId `json:"-" bson:"_id,omitempty"`
	GUID      MongoUUID     `json:"guid" bson:"guid"`
	Timestamp time.Time     `json:"timestamp" bson:"timestamp"`
	Status    Status        `json:"status" bson:"status"`
}

func newTask() *Task {
	return &Task{
		GUID:      MongoUUID{uuid.NewV4()},
		Timestamp: time.Now(),
		Status:    Created,
	}

}

func findTaskByGUID(db *DB, id uuid.UUID) (*Task, error) {
	task := new(Task)
	col := db.Session.DB(config.MongoDatabase).C(config.Collection)
	err := col.Find(bson.M{"guid": bson.Binary{
		Kind: bson.BinaryUUID,
		Data: id.Bytes(),
	}}).One(task)
	return task, err
}

func updateTaskByGUID(db *DB, id uuid.UUID, update bson.M) error {
	col := db.Session.DB(config.MongoDatabase).C(config.Collection)
	colQuerier := bson.M{"guid": bson.Binary{
		Kind: bson.BinaryUUID,
		Data: id.Bytes(),
	}}
	change := bson.M{"$set": update}
	err := col.Update(colQuerier, change)
	return err
}

type TaskRunner struct {
	TaskExecutionTimeSeconds int
	DB                       *DB
}

func newTaskRunner(cfg *Config, db *DB) *TaskRunner {
	return &TaskRunner{
		TaskExecutionTimeSeconds: cfg.TaskExecutionTimeSeconds,
		DB:                       db,
	}
}

func (tr TaskRunner) run(id uuid.UUID) {
	err := updateTaskByGUID(tr.DB, id, bson.M{"timestamp": time.Now(), "status": Running})
	if err != nil {
		log.Println(err)
		return
	}

	time.Sleep(time.Duration(tr.TaskExecutionTimeSeconds) * time.Second)

	err = updateTaskByGUID(tr.DB, id, bson.M{"timestamp": time.Now(), "status": Finished})
	if err != nil {
		log.Println(err)
		return
	}
}

type MongoUUID struct{ uuid.UUID }

func (id *MongoUUID) SetBSON(raw bson.Raw) error {
	var guid bson.Binary
	err := raw.Unmarshal(&guid)
	if err != nil {
		return err
	}

	uu, err := uuid.FromBytes(guid.Data)
	if err != nil {
		return err
	}

	*id = MongoUUID{uu}

	return nil
}

func (id MongoUUID) GetBSON() (interface{}, error) {
	ret := bson.Binary{
		Kind: bson.BinaryUUID,
		Data: nil,
	}

	if uuid.Equal(id.UUID, uuid.Nil) {
		ret.Data = uuid.Nil.Bytes()
	} else {
		ret.Data = id.Bytes()
	}

	return ret, nil
}
