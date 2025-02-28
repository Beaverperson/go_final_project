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
	fmt.Printf("DEBUG API nextdate handler GET param \"now\": %s\n", nowParam)
	fmt.Printf("DEBUG API nextdate handler GET param \"date\": %s\n", dateParam)
	fmt.Printf("DEBUG API nextdate handler GET param \"repeat\": %s\n", repeatParam)
	if len(nowParam) == 0 || len(dateParam) == 0 {
		fmt.Print("ERROR API incorrect date parameters /api/nextdate\n")
		http.Error(w, "incorrect date parameters /api/nextdate\n", http.StatusBadRequest)
	}
	w.Header().Set("Content-Type", "text/html")
	nowTime, err := time.Parse("20060102", nowParam)
	if err != nil {
		fmt.Print("ERROR API can't convert \"now\" to time format\n")
		http.Error(w, "can't convert \"now\" to time format\n", http.StatusBadRequest)
	}
	nextDateResponce, err := NextDate(nowTime, dateParam, repeatParam)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	w.Write([]byte(nextDateResponce))
}
