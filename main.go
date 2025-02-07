package main

import (
	"fmt"
	"net/http"

	"github.com/Beaverperson/go_final_project/tests"
)

func main() {
	//
	http.Handle("/", http.FileServer(http.Dir(tests.WebDir)))
	err := http.ListenAndServe(fmt.Sprintf(":%d", tests.Port), nil)
	if err != nil {
		fmt.Printf("ошибка запуска сервера: %s\n", err.Error())
		return
	}
}
