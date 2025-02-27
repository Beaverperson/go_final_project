package main

import (
	"fmt"
	"net/http"
)

const (
	webDir  = "./web"
	webPort = "7540"
	dbName  = "scheduler.db"
	SQLinit = `CREATE TABLE IF NOT EXISTS scheduler (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		date TEXT,
		title TEXT,
		comment TEXT,
		repeat TEXT);
		CREATE INDEX IF NOT EXISTS indexdate ON scheduler (date);`
)

func main() {

	http.Handle("/", http.FileServer(http.Dir(webDir)))

	fmt.Printf("INFO starting web server on port:%s\n", webPort)
	err := http.ListenAndServe(":"+webPort, nil)
	if err != nil {
		fmt.Printf("ERROR web server isn't started: %s", err)
	}
}
