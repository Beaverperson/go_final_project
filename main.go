package main

import (
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
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
	db, err := GetDBConnector()
	if err != nil {
		log.Fatal()
	}
	defer db.Close()

	// WEB
	//WebPort := getenv("TODO_PORT", strconv.Itoa(tests.Port))
	//http.Handle("/", http.FileServer(http.Dir(tests.WebDir)))
	//if err := http.ListenAndServe(":"+WebPort, nil); err != nil {
	//	fmt.Printf("ошибка запуска сервера: %s\n", err.Error())
	//	return
	//}
}
