package main

import (
	"github.com/jinzhu/configor"
)

var config *Config

func main() {
	configor.Load(&config, "config.yml")
	a := App{}
	a.Initialize(config)
	a.Run(config)
}
