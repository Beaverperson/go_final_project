package main

import (
	"database/sql"
	"encoding/json"
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
		http.Error(w, "incorrect date parameters /api/nextdate\n",
			http.StatusBadRequest)
	}
	w.Header().Set("Content-Type", "text/html")
	nowTime, err := time.Parse(dateFormat, nowParam)
	if err != nil {
		fmt.Print("ERROR API can't convert \"now\" to time format\n")
		http.Error(w, "can't convert \"now\" to time format\n",
			http.StatusBadRequest)
	}
	nextDateResponce, err := NextDate(nowTime, dateParam, repeatParam)
	if err != nil {
		http.Error(w, err.Error(),
			http.StatusBadRequest)
	}
	w.Write([]byte(nextDateResponce))
}

func HandlerAPITask(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var task Task
	switch {
	case r.Method == http.MethodPost:
		fmt.Print("INFO API task received POST message \"/api/task\"\n")
		err := json.NewDecoder(r.Body).Decode(&task)
		if err != nil {
			fmt.Print("ERROR API unable to deserialize JSON  \"/api/task\"\n")
			http.Error(w, `{"error": "JSON deserialization"}`,
				http.StatusBadRequest)
		}
		if task.Title == "" {
			fmt.Print("ERROR API title is mandatory  \"/api/task\"\n")
			http.Error(w, `{"error": "Title is required"}`,
				http.StatusBadRequest)
			return
		}
		if task.Date == "" {
			fmt.Print("DEBUG API task date is missing  \"/api/task\"\n")
			task.Date = time.Now().Format(dateFormat)
		} else {
			parsedDate, err := time.Parse(dateFormat, task.Date)
			if err != nil {
				http.Error(w, `{"error": "Invalid date format"}`,
					http.StatusBadRequest)
				return
			}
			if parsedDate.Before(time.Now()) {
				if task.Repeat == "" {
					task.Date = time.Now().Format(dateFormat)
				} else {
					nextDate, err := NextDate(time.Now(), task.Date, task.Repeat)
					if err != nil {
						http.Error(w, fmt.Sprintf(`{"error": "Invalid repeat rule: %s"}`, err),
							http.StatusBadRequest)
						return
					}
					task.Date = nextDate
				}
			}
		}
		query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`
		fmt.Printf("INFO SQL attempt to submit data to scheduler (%s, %s, %s, %s)\n",
			task.Date, task.Title, task.Comment, task.Repeat)
		result, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
		if err != nil {
			fmt.Printf("ERROR SQL unable to submit data to scheduler (%s, %s, %s, %s)\n",
				task.Date, task.Title, task.Comment, task.Repeat)
			http.Error(w, fmt.Sprintf(`{"error": "Unable to submit data to DB: %v"}`, err),
				http.StatusInternalServerError)
			return
		}
		id, err := result.LastInsertId()
		if err != nil {
			fmt.Print("ERROR SQL unable to get last scheduler task ID")
			http.Error(w, fmt.Sprintf(`{"error": "Unable to get last scheduler task ID: %s"}`, err),
				http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"id": id})
	default:
		fmt.Print("ERROR API wrong http method while accessing \"/api/task\"\n")
		http.Error(w, fmt.Sprintf("wrong http method: %s\n", r.Method), http.StatusMethodNotAllowed)
		return
	}
}

func HandlerAPITaskS(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	fmt.Printf("INFO API received get tasks message \"/api/task\"\n")
	tasks := []Task{}
	count := 0
	rowsCount := db.QueryRow("SELECT count(*) FROM scheduler")
	_ = rowsCount.Scan(&count)
	fmt.Printf("INFO SQL number of tasks in scheduler: %d", count)
	if count > 0 {
		query := fmt.Sprintf(`SELECT TOP %d FROM scheduler ORDER BY date`, maxRowsTasks)
		rowsData, err := db.Query(query)
		if err != nil {
			fmt.Print("ERROR SQL unable to get current tasks")
			http.Error(w, fmt.Sprintf(`{"error": "Unable to get current tasks: %s"}`, err),
				http.StatusInternalServerError)
			return
		}
		fmt.Print("INFO SQL get task(s) from scheduler")
		var task Task
		for rowsData.Next() {
			err := rowsData.Scan(
				&task.ID,
				&task.Date,
				&task.Title,
				&task.Comment,
				&task.Repeat)
			if err != nil {
				fmt.Print("ERROR SQL unable to parce current tasks")
				http.Error(w, fmt.Sprintf(`{"error": "Unable to parce current tasks: %s"}`, err),
					http.StatusInternalServerError)
			}
			tasks = append(tasks, task)
		}
		fmt.Printf("INFO SQL received %d task(s) from scheduler DB", len(tasks))
		resp, err := json.Marshal(tasks)
		if err != nil {
			fmt.Print("ERROR API unable to seriliaze current tasks")
			http.Error(w, fmt.Sprintf(`{"error": "Unable to seriliaze current tasks: %s"}`, err),
				http.StatusInternalServerError)
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	}
	if count == 0 {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"tasks": tasks})
	}
}
