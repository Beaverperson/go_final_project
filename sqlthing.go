package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

func GetDBConnector() (*sql.DB, error) {
	appPath, err := os.Getwd()
	if err != nil {
		fmt.Printf("Не получилось найти текущую рабочую директорию. с каждым могло произойти")
		return nil, err
	}
	dbFileName := getenv("TODO_DBFILE", "scheduler.db")
	fmt.Printf("Имя файла для базы: %s\n", dbFileName)
	dbFile := filepath.Join(appPath, dbFileName)
	fmt.Printf("Полный путь к файлу базы: %s\n", dbFile)
	_, err = os.Stat(dbFile)
	if err != nil {
		os.Create(dbFile)
		fmt.Printf("Новая база \"%s\" создана. Радостно", dbFileName)
	}
	dbCreator, errOpen := sql.Open("sqlite3", dbFile)
	if errOpen != nil {
		return nil, fmt.Errorf("ошибка при открытии \"%s\"", dbFileName)
	}
	fmt.Print("Подключение к базе для проверки наличия стартовой таблицы\n")
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
	return dbCreator, nil
}
