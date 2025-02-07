package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

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
	WebPort := getenv("TODO_PORT", strconv.Itoa(tests.Port))
	http.Handle("/", http.FileServer(http.Dir(tests.WebDir)))
	err := http.ListenAndServe(":"+WebPort, nil)
	if err != nil {
		fmt.Printf("ошибка запуска сервера: %s\n", err.Error())
		return
	}
}
