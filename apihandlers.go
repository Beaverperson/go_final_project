package main

import (
	"fmt"
	"net/http"
	"time"
)

func nextDateHandler(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	now := params.Get("now")
	date := params.Get("date")
	repeat := params.Get("repeat")
	if len(now) == 0 || len(date) == 0 {
		fmt.Print("Некорретный формат параметров в запросе к /api/nextdate\n")
		http.Error(w, "Некорретный формат параметров в запросе к /api/nextdate\n", http.StatusBadRequest)
	}

	w.Header().Set("Content-Type", "text/html")
	nowTime, err := time.Parse("20250301", now)
	if err != nil {
		fmt.Print("значение \"now\" не конвертируется в дату\n")
		http.Error(w, "значение \"now\" не конвертируется в дату\n", http.StatusBadRequest)
	}
	nextDateResponce, err := NextDate(nowTime, date, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	w.Write([]byte(nextDateResponce))
}
