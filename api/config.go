package main

type Config struct {
	Addr                     string `default:":80"`
	MongoHosts               string `default:"localhost:27017"`
	MongoDatabase            string `default:"db"`
	MongoUsername            string `default:""`
	MongoPassword            string `default:""`
	Collection               string `default:"tasks"`
	TaskExecutionTimeSeconds int    `default:"120"`
}
