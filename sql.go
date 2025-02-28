package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

func GetDBConnector(dbFileName string) (*sql.DB, error) {
	appPath, err := os.Getwd()
	if err != nil {
		fmt.Printf("ERROR SQL unable to find working directory\n")
		return nil, err
	}
	dbFile := filepath.Join(appPath, dbFileName)
	fmt.Printf("INFO SQL full path to DB file: %s", dbFile)
	_, err = os.Stat(dbFile)
	if err != nil {
		fmt.Print("INFO SQL DB is missing. creating... \n")
		os.Create(dbFile)
	}
	dbCreator, errOpen := sql.Open("sqlite3", dbFile)
	if errOpen != nil {
		return nil, fmt.Errorf("unable to open DB \"%s\"", dbFileName)
	}
	fmt.Print("\n")
	_, errCreate := dbCreator.Exec(`
		CREATE TABLE IF NOT EXISTS scheduler (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		date REAL NOT NULL,
		title TEXT,
		comment	TEXT,
		repeat TEXT
		);
		CREATE INDEX IF NOT EXISTS indexdate ON scheduler (date);
		`)
	if errCreate != nil {
		return nil, errCreate
	}
	fmt.Printf("INFO SQL DB \"%s\" id ready to use\n", dbFile)
	return dbCreator, nil
}
