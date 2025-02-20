package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func getDBConnector() (*sql.DB, error) {
	appPath, err := os.Getwd()
	if err != nil {
		fmt.Printf("Не получилось найти текущую рабочую директорию. с каждым могло произойти")
		return nil, err
	}
	dbFileName := getenv("TODO_DBFILE", "scheduler.db")
	dbFile := filepath.Join(appPath, dbFileName)
	_, err = os.Stat(dbFile)
	if err == nil {
		dbCreator, errOpen := sql.Open("sqlite3", dbFile)
		if errOpen != nil {
			return nil, fmt.Errorf("ошибка при открытии \"%s\"", dbFileName)
		}
		return dbCreator, nil
	}
	os.Create(dbFile)
	dbCreator, errOpen := sql.Open("sqlite3", dbFile)
	if errOpen != nil {
		return nil, fmt.Errorf("ошибка при открытии \"%s\"", dbFileName)
	}
	if errOpen != nil {
		_, errCreate := dbCreator.Exec(`CREATE TABLE IF NOT EXIST "scheduler" (
											"id"	INTEGER NOT NULL,
											"date"	REAL NOT NULL,
											"title"	TEXT,
											"comment"	TEXT,
											"repeat"	TEXT,
											PRIMARY KEY("id" AUTOINCREMENT)
										);
										CREATE INDEX IF NOT EXIST indexdate ON scheduler (date);`)
		fmt.Printf("Новая база \"%s\" создана", dbFileName)
		if errCreate != nil {
			return nil, errCreate
		}
	} else {
		return nil, errOpen
	}
	return dbCreator, nil
}
