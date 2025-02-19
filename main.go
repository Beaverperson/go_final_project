package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	_ "github.com/mattn/go-sqlite3"

	"github.com/Beaverperson/go_final_project/tests"
)

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func main() {
	// setup SQL
	appPath, err := os.Getwd()
	if err != nil {
		fmt.Printf("Не получилось найти текущую рабочую директорию. с каждым могло произойти")
	}
	dbFileName := getenv("TODO_DBFILE", "scheduler.db")
	dbFile := filepath.Join(appPath, dbFileName)
	_, err = os.Stat(dbFile)
	// if DB file is missing - create
	if err != nil {
		os.Create(dbFile)
		dbCreator, errOpen := sql.Open("sqlite3", dbFile)
		if errOpen == nil {
			_, errCreate := dbCreator.Exec(`CREATE TABLE "scheduler" (
												"id"	INTEGER NOT NULL,
												"date"	REAL NOT NULL,
												"title"	TEXT,
												"comment"	TEXT,
												"repeat"	TEXT,
												PRIMARY KEY("id" AUTOINCREMENT)
											);
											CREATE INDEX indexdate ON scheduler (date);`)
			fmt.Printf("Новая база \"%s\" создана", dbFileName)
			if errCreate != nil {
				fmt.Printf("Ошибка записи в создаваемой базе: \"%s\"\n", errCreate.Error())
			}
			dbCreator.Close()
		} else {
			fmt.Printf("Ошибка доступа к создаваемой базе: \"%s\"\n", errOpen.Error())
		}
	} else {
		fmt.Printf("База \"%s\" уже существует\n", dbFileName)
	}
	// WEB
	WebPort := getenv("TODO_PORT", strconv.Itoa(tests.Port))
	http.Handle("/", http.FileServer(http.Dir(tests.WebDir)))
	if err := http.ListenAndServe(":"+WebPort, nil); err != nil {
		fmt.Printf("ошибка запуска сервера: %s\n", err.Error())
		return
	}
}
