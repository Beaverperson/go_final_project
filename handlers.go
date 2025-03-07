package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
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
		fmt.Printf("DEBUG API POST message \"/api/task\" serialization (id,%s,%s,%s,%s)\n",
			task.Date, task.Title, task.Comment, task.Repeat)
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
			//if parsedDate.Before(time.Now()) {
			if parsedDate.Format(dateFormat) < time.Now().Format(dateFormat) {
				fmt.Print("DEBUG API task date from \"/api/task\" is BEFORE NOW\n")
				fmt.Printf("Parsed: %s, Now: %s\n", parsedDate.Format(dateFormat), time.Now().Format(dateFormat))
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
			http.Error(w, fmt.Sprintf(`{"error":"Unable to submit data to DB: %v"}`, err),
				http.StatusInternalServerError)
			return
		}
		id, err := result.LastInsertId()
		if err != nil {
			fmt.Print("ERROR SQL unable to get last scheduler task ID")
			http.Error(w, fmt.Sprintf(`{"error":"Unable to get last scheduler task ID: %s"}`, err),
				http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"id": id})
	case r.Method == http.MethodGet:
		fmt.Print("INFO API task received GET message \"/api/task\"\n")
		taskId := r.URL.Query().Get("id")
		//fmt.Print(r.URL)
		fmt.Printf("DEBUG API \"/api/task\" GET param \"id\": %s\n", taskId)
		if len(taskId) == 0 {
			fmt.Print("ERROR API incorrect task id parameter \"/api/task\"\n")
			http.Error(w, `{"error":"Task ID is missing"}`,
				http.StatusBadRequest)
			return
		}
		query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`
		row := db.QueryRow(query, taskId)
		err := row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err == sql.ErrNoRows {
			fmt.Printf("INFO SQL task with ID %s not found\n", taskId)
			http.Error(w, `{"error": "Task with ID not found"}`,
				http.StatusBadRequest)
			return
		} else if err != nil {
			fmt.Printf("ERROR SQL error while retrieving task with ID %s: %s\n", taskId, err)
			http.Error(w, `{"error": "Task ID is missing"}`,
				http.StatusInternalServerError)
			return
		}
		resp, err := json.Marshal(task)
		if err != nil {
			fmt.Print("ERROR API unable to seriliaze selected task")
			http.Error(w, fmt.Sprintf(`{"error": "Unable to seriliaze current task %s:"}`, taskId),
				http.StatusInternalServerError)
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	case r.Method == http.MethodPut:
		fmt.Print("INFO API task received PUT message \"/api/task\"\n")
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			fmt.Print("ERROR API unable to deserialize JSON")
			http.Error(w, `{"error":"Unable to deserialize JSON"}`, http.StatusBadRequest)
			return
		}
		fmt.Printf("DEBUG API PUT message \"/api/task\" %+v\n", task)
		if _, err := strconv.Atoi(task.ID); err != nil {
			//TODO надо бы в легулярку чтобы не подгружать библиотеку strconv
			fmt.Printf("ERROR API ID is not a number: %s\n", task.ID)
			http.Error(w, `{"error":"ID is not a number"}`, http.StatusBadRequest)
			return
		}
		if task.ID == "" || task.Title == "" {
			fmt.Print("ERROR API missing mandatory fields in PUT message\n")
			http.Error(w, `{"error":"Missing mandatory fields in PUT message"}`, http.StatusBadRequest)
			return
		}
		if _, err := time.Parse(dateFormat, task.Date); err != nil {
			fmt.Print("ERROR API unable to format date in PUT message\n")
			http.Error(w, `{"error":"Unable to format date in PUT message"}`, http.StatusBadRequest)
			return
		}
		if (task.Repeat != "" && task.Repeat != "y") &&
			!regexp.MustCompile(daysRegex).MatchString(task.Repeat) {
			fmt.Print("ERROR API unable to format repeat in PUT message\n")
			http.Error(w, `{"error":"Unable to format repeat in PUT message"}`, http.StatusBadRequest)
			return
		}
		query := `UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`
		execResult, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
		if err != nil {
			fmt.Printf("ERROR SQL unable to update task with ID: %s\n", task.ID)
			http.Error(w, `{"error":"Unable to format repeat in PUT message"}`, http.StatusInternalServerError)
			return
		}
		rowsAffected, err := execResult.RowsAffected()
		if err != nil {
			fmt.Printf("ERROR SQL didn't reiceved affected rows from DB after updating task ID: %s\n", task.ID)
			http.Error(w, `{"error":"Unable to format repeat in PUT message"}`, http.StatusInternalServerError)
			return
		}
		if rowsAffected == 0 {
			fmt.Printf("INFO SQL there is no task with ID (update failed): %s\n", task.ID)
			http.Error(w, `{"error":"Corresponding ID not found in SQL database"}`, http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{})
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
	fmt.Printf("INFO SQL number of tasks in scheduler: %d\n", count)
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
