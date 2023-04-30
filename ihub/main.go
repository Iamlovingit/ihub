package main

import (
	"ihub/pkg/config"
	"ihub/pkg/db"

	"ihub/pkg/server"
)

func main() {
	//todo use cache to instead of config and db
	//init cofig from file or environment
	if err := config.Init("ihub-config.yaml"); err != nil {
		panic(err)
	}

	//init database to get db handler
	if err := db.Init(); err != nil {
		panic(err)
	}

	//init server
	app := server.NewServer()
	if app == nil {
		panic("New Server failed.")
	}
	app.Run()
}
