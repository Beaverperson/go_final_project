package main

import (
	"fmt"
	"net/http"
	"time"
)

func HandlerNextDate(w http.ResponseWriter, r *http.Request) {
	nowParam := r.FormValue("now")
	dateParam := r.FormValue("date")
	repeatParam := r.FormValue("repeat")
	fmt.Printf("DEBUG nextdate handler GET param \"now\": %s\n", nowParam)
	fmt.Printf("DEBUG nextdate handler GET param \"date\": %s\n", dateParam)
	fmt.Printf("DEBUG nextdate handler GET param \"repeat\": %s\n", repeatParam)
	if len(nowParam) == 0 || len(dateParam) == 0 {
		fmt.Print("Некорретный формат параметров в запросе к /api/nextdate\n")
		http.Error(w, "Некорретный формат параметров в запросе к /api/nextdate\n", http.StatusBadRequest)
	}
	w.Header().Set("Content-Type", "text/html")
	nowTime, err := time.Parse("20060102", nowParam)
	if err != nil {
		fmt.Print("значение \"now\" не конвертируется в дату\n")
		http.Error(w, "значение \"now\" не конвертируется в дату\n", http.StatusBadRequest)
	}
	nextDateResponce, err := NextDate(nowTime, dateParam, repeatParam)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	w.Write([]byte(nextDateResponce))
}
