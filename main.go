package main

import (
	log "main/logging"
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
	log.Info("ROOT establishing connection to DB:%s", dbName)
	if err != nil {
		log.Fatal("ROOT DB is unavailable. Terminating: (%s)", err.Error())
	}
	defer db.Close()
	// START WEB
	webPort := Getenv("TODO_PORT", webPortDefault)
	log.Info("ROOT starting web server on port:%s", webPort)
	// HANDLERS
	http.Handle("/", http.FileServer(http.Dir(webDir)))
	http.HandleFunc("/api/nextdate", HandlerNextDate)
	http.HandleFunc("/api/task", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			HandlerAddTask(w, r, db)
		case http.MethodGet:
			HandlerGetTask(w, r, db)
		case http.MethodPut:
			HandlerUpdateTask(w, r, db)
		case http.MethodDelete:
			HandlerDeleteTask(w, r, db)
		default:
			HandlerOTHER(w, r)
		}
	})
	http.HandleFunc("/api/tasks", func(w http.ResponseWriter, r *http.Request) {
		HandlerAPITaskS(w, r, db)
	})
	http.HandleFunc("/api/task/done", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			HandlerAPITaskDone(w, r, db)
		case http.MethodDelete:
			HandlerDeleteTask(w, r, db)
		default:
			HandlerOTHER(w, r)
		}
	})
	if http.ListenAndServe(":"+webPort, nil) != nil {
		log.Fatal("ROOT web server isn't started. Terminating: (%s)", err.Error())
	}
}
