package main

import (
	"fmt"
	"log"
	"net/http"
)

const (
	webDir         = "./web"
	webPortDefault = "7540"
	dbNameDefault  = "scheduler.db"
	SQLinit        = `CREATE TABLE IF NOT EXISTS scheduler (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            date TEXT NOT NULL,
            title TEXT NOT NULL,
            comment TEXT,
            repeat TEXT
			);
		CREATE INDEX IF NOT EXISTS indexdate ON scheduler (date);`
	dateFormat   = "20060102"
	maxRowsTasks = 10
	daysRegex    = `^d\s(\d+)$`
)

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment,omitempty"`
	Repeat  string `json:"repeat,omitempty"`
}

func main() {
	// START SQL DB
	dbName := Getenv("TODO_DBFILE", dbNameDefault)
	db, err := GetDBConnector(dbName)
	fmt.Printf("INFO ROOT establishing connection to DB:%s\n", dbName)
	if err != nil {
		fmt.Printf("ERROR ROOT DB is unavailable. Terminating: %s\n", err.Error())
		log.Fatal()
	}
	defer db.Close()
	// START WEB
	webPort := Getenv("TODO_PORT", webPortDefault)
	fmt.Printf("INFO ROOT starting web server on port:%s\n", webPort)
	// HANDLERS
	http.Handle("/", http.FileServer(http.Dir(webDir)))
	http.HandleFunc("/api/nextdate", HandlerNextDate)
	http.HandleFunc("/api/task", func(w http.ResponseWriter, r *http.Request) {
		HandlerAPITask(w, r, db)
	})
	http.HandleFunc("/api/tasks", func(w http.ResponseWriter, r *http.Request) {
		HandlerAPITaskS(w, r, db)
	})
	http.HandleFunc("/api/task/done", func(w http.ResponseWriter, r *http.Request) {
		HandlerAPITaskDone(w, r, db)
	})
	//http.HandleFunc("/", HandlerOTHER)
	if http.ListenAndServe(":"+webPort, nil) != nil {
		fmt.Printf("ERROR ROOT web server isn't started: %s\n", err.Error())
		log.Fatal()
	}
}
