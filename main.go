package main

import (
	"fmt"
	_ "github.com/joho/godotenv/autoload"
	"github.com/nezuchan/scheduled-tasks/config"
	"github.com/nezuchan/scheduled-tasks/lib"
)

func main() {
	conf, err := config.Init()
	if err != nil {
		panic(fmt.Sprintf("couldn't initialize config: %v", err))
	}
	_ = lib.InitTask(&conf)

	select {}
}
