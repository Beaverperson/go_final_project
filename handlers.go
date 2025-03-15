package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"

	log "main/logging"
)

func HandlerNextDate(w http.ResponseWriter, r *http.Request) {
	nowParam := r.FormValue("now")
	dateParam := r.FormValue("date")
	repeatParam := r.FormValue("repeat")
	log.Debug("API nextdate handler GET param 'now': %s", nowParam)
	log.Debug("API nextdate handler GET param 'date': %s", dateParam)
	log.Debug("API nextdate handler GET param 'repeat': %s", repeatParam)
	if len(nowParam) == 0 || len(dateParam) == 0 {
		log.Error("ERROR API incorrect date parameters /api/nextdate %s", dateParam)
		http.Error(w, "incorrect date parameters /api/nextdate",
			http.StatusBadRequest)
	}
	w.Header().Set("Content-Type", "text/html")
	nowTime, err := time.Parse(dateFormat, nowParam)
	if err != nil {
		log.Error("API can't convert 'now' to time format (%s)", err.Error())
		http.Error(w, "can't convert 'now' to time format",
			http.StatusBadRequest)
	}
	nextDateResponce, err := NextDate(nowTime, dateParam, repeatParam)
	if err != nil {
		http.Error(w, err.Error(),
			http.StatusBadRequest)
	}
	w.Write([]byte(nextDateResponce))
}

func HandlerAddTask(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var task Task
	log.Info("API task received POST message '/api/task'")
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		log.Error("API unable to deserialize JSON '/api/task' (%s)", err.Error())
		http.Error(w, `{"error": "JSON deserialization"}`,
			http.StatusBadRequest)
	}
	log.Debug("API POST message '/api/task' serialization %v",
		task)
	if task.Title == "" {
		log.Error("API title is mandatory '/api/task' (%s)", err.Error())
		http.Error(w, `{"error": "Title is required"}`,
			http.StatusBadRequest)
		return
	}
	if task.Date == "" {
		log.Debug("API task date is missing '/api/task' (%s)", err.Error())
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
			log.Debug("DEBUG API task date from '/api/task' is BEFORE NOW")
			log.Info("Parsed: %s, Now: %s", parsedDate.Format(dateFormat), time.Now().Format(dateFormat))
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
	log.Info("SQL attempt to submit data to scheduler %v", task)
	result, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		log.Error("SQL unable to submit data to scheduler %v (%s)", task, err.Error())
		http.Error(w, fmt.Sprintf(`{"error":"Unable to submit data to DB: %v"}`, err),
			http.StatusInternalServerError)
		return
	}
	id, err := result.LastInsertId()
	if err != nil {
		log.Error("SQL unable to get last scheduler task ID (%s)", err.Error())
		http.Error(w, fmt.Sprintf(`{"error":"Unable to get last scheduler task ID: %s"}`, err),
			http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"id": id})
}

func HandlerGetTask(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var task Task
	log.Info("API task received GET message '/api/task'")
	taskId := r.URL.Query().Get("id")
	log.Debug("API '/api/task' GET param 'id': %s", taskId)
	if len(taskId) == 0 {
		log.Error("API incorrect task id parameter '/api/task'")
		http.Error(w, `{"error":"Task ID is missing"}`,
			http.StatusBadRequest)
		return
	}
	query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`
	row := db.QueryRow(query, taskId)
	err := row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err == sql.ErrNoRows {
		log.Info("SQL task with ID %s not found", taskId)
		http.Error(w, `{"error": "Task with ID not found"}`,
			http.StatusBadRequest)
		return
	} else if err != nil {
		log.Error("SQL error while retrieving task with ID %s (%s)", taskId, err.Error())
		http.Error(w, `{"error": "Task ID is missing"}`,
			http.StatusInternalServerError)
		return
	}
	resp, err := json.Marshal(task)
	if err != nil {
		log.Error("API unable to seriliaze selected task ()", err.Error())
		http.Error(w, fmt.Sprintf(`{"error": "Unable to seriliaze current task %s:"}`, taskId),
			http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func HandlerUpdateTask(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var task Task
	log.Info("API task received PUT message '/api/task'")
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		log.Error("API unable to deserialize JSON (%s)", err.Error())
		http.Error(w, `{"error":"Unable to deserialize JSON"}`, http.StatusBadRequest)
		return
	}
	log.Debug("API PUT message '/api/task' %+v", task)
	if _, err := strconv.Atoi(task.ID); err != nil {
		//TODO надо бы в легулярку чтобы не подгружать библиотеку strconv
		log.Error("API ID is not a number - [%s] (%s)", task.ID, err.Error())
		http.Error(w, `{"error":"ID is not a number"}`, http.StatusBadRequest)
		return
	}
	if task.ID == "" || task.Title == "" {
		log.Info("API missing mandatory fields in PUT message")
		http.Error(w, `{"error":"Missing mandatory fields in PUT message"}`, http.StatusBadRequest)
		return
	}
	if _, err := time.Parse(dateFormat, task.Date); err != nil {
		log.Error("API unable to format date in PUT message (%s)", err.Error())
		http.Error(w, `{"error":"Unable to format date in PUT message"}`, http.StatusBadRequest)
		return
	}
	if (task.Repeat != "" && task.Repeat != "y") &&
		!regexp.MustCompile(daysRegex).MatchString(task.Repeat) {
		log.Error("API unable to format repeat in PUT message")
		http.Error(w, `{"error":"Unable to format repeat in PUT message"}`, http.StatusBadRequest)
		return
	}
	query := `UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`
	execResult, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		log.Error("SQL unable to update task with ID %s (%s)", task.ID, err.Error())
		http.Error(w, `{"error":"SQL exec failure"}`, http.StatusInternalServerError)
		return
	}
	rowsAffected, err := execResult.RowsAffected()
	if err != nil {
		log.Error("SQL didn't reiceved affected rows from DB after updating task ID %s (%s)", task.ID, err.Error())
		http.Error(w, `{"error":"Unable to format repeat in PUT message"}`, http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		log.Info("SQL there is no task with ID (update failed): %s", task.ID)
		http.Error(w, `{"error":"Corresponding ID not found in SQL database"}`, http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{})
}

func HandlerDeleteTask(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var task Task
	taskId := r.URL.Query().Get("id")
	if taskId == "" {
		log.Error("API task ID field is empty")
		http.Error(w, `{"error":"task ID field is empty"}`, http.StatusBadRequest)
		return
	}
	if _, err := strconv.Atoi(taskId); err != nil {
		//TODO надо бы в легулярку чтобы не подгружать библиотеку strconv
		log.Error("API ID is not a number %s (%s)", taskId, err.Error())
		http.Error(w, `{"error":"ID is not a number"}`, http.StatusBadRequest)
		return
	}
	log.Info("API received DELETE message '/api/task'")
	query := `DELETE FROM scheduler WHERE id = ?`
	execResult, err := db.Exec(query, taskId)
	if err != nil {
		log.Error("SQL unable to delete task with ID %s (%s)", taskId, err.Error())
		http.Error(w, `{"error":"SQL exec failure"}`, http.StatusInternalServerError)
		return
	}
	rowsAffected, err := execResult.RowsAffected()
	if err != nil {
		log.Error("SQL didn't reiceved affected rows from DB after deleting task ID %s (%s)", task.ID, err.Error())
		http.Error(w, `{"error":"Unable to format repeat in PUT message"}`, http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		log.Info("SQL there is no task with ID (delete failed): %s", task.ID)
		http.Error(w, `{"error":"Corresponding ID not found in SQL database"}`, http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{})
}

func HandlerDoneTask(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var task Task
	taskId := r.URL.Query().Get("id")
	if taskId == "" {
		log.Error("API task ID field is empty")
		http.Error(w, `{"error":"task ID field is empty"}`, http.StatusBadRequest)
	}
	if _, err := strconv.Atoi(taskId); err != nil {
		log.Error("API ID is not a number %s (%s)", taskId, err.Error())
		http.Error(w, `{"error":"ID is not a number"}`, http.StatusBadRequest)
		return
	}
	log.Info("API received POST message '/api/task/done'")
	query := `SELECT id, date, repeat FROM scheduler WHERE id = ?`
	err := db.QueryRow(query, taskId).Scan(&task.ID, &task.Date, &task.Repeat)
	if err == sql.ErrNoRows {
		log.Info("SQL there is no task with ID %s (update failed):", taskId)
		http.Error(w, `{"error":"Corresponding ID not found in SQL database"}`, http.StatusNotFound)
		return
	} else if err != nil {
		log.Error("SQL unable to update task with ID %s (%s)", task.ID, err.Error())
		http.Error(w, `{"error":"SQL exec failure"}`, http.StatusInternalServerError)
		return
	}
	log.Debug("API '/api/task/done' retrived from DB %+v", task)
	if task.Repeat == "" {
		log.Debug("SQL 'repeat' is empty, trying to delete task with id: %s", task.ID)
		_, err := db.Exec(`DELETE FROM scheduler WHERE id = ?`, taskId)
		if err != nil {
			log.Error("SQL unable to delete task with ID %s (%s)", task.ID, err.Error())
			http.Error(w, `{"error":"SQL exec failure"}`, http.StatusInternalServerError)
			return
		}
	} else {
		log.Debug("API 'repeat' is not empty, trying to calculate new date for task ID: %s", task.ID)
		now := time.Now()
		nextDate, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			log.Debug("API failed to calculate new date for task ID: %s (%s)", task.ID, err.Error())
			http.Error(w, `{"error":"failed to calculate new date for the task"}`, http.StatusInternalServerError)
			return
		}
		log.Debug("API new date calculated, proceed to update DB")
		_, err = db.Exec(`UPDATE scheduler SET date = ? WHERE id = ?`, nextDate, taskId)
		if err != nil {
			log.Error("SQL failed to UPDATE ID: %s (%s)", task.ID, err.Error())
			http.Error(w, `{"error":"SQL exec failure"}`, http.StatusInternalServerError)
			return
		}
		log.Info("SQL task with ID: %s updated", taskId)
	}
	json.NewEncoder(w).Encode(map[string]string{})
}

func HandlerAPITaskS(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	log.Info("API received get tasks message '/api/task'")
	tasks := []Task{}
	count := 0
	rowsCount := db.QueryRow("SELECT count(*) FROM scheduler")
	_ = rowsCount.Scan(&count)
	log.Info("SQL number of tasks in scheduler: %d", count)
	if count == 0 {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"tasks": tasks})
		return
	}
	query := `SELECT * FROM scheduler ORDER BY date ASC LIMIT ?`
	rowsData, err := db.Query(query, maxRowsTasks)
	if err != nil {
		log.Error("SQL unable to get current tasks (%s)", err.Error())
		http.Error(w, fmt.Sprintf(`{"error": "Unable to get current tasks: %s"}`, err),
			http.StatusInternalServerError)
		return
	}
	log.Info("SQL get task(s) from scheduler")
	for rowsData.Next() {
		var task Task
		err := rowsData.Scan(
			&task.ID,
			&task.Date,
			&task.Title,
			&task.Comment,
			&task.Repeat)
		if err != nil {
			log.Error("SQL unable to parce current tasks (%s)", err.Error())
			http.Error(w, fmt.Sprintf(`{"error": "Unable to parce current tasks: %s"}`, err),
				http.StatusInternalServerError)
		}
		tasks = append(tasks, task)
	}
	log.Info("SQL received %d task(s) from scheduler DB", len(tasks))
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"tasks": tasks})
}

func HandlerAPITaskDone(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var task Task
	taskId := r.URL.Query().Get("id")
	if taskId == "" {
		log.Error("API task ID field is empty")
		http.Error(w, `{"error":"task ID field is empty"}`, http.StatusBadRequest)
	}
	if _, err := strconv.Atoi(taskId); err != nil {
		log.Error("API ID is not a number %s (%s)", taskId, err.Error())
		http.Error(w, `{"error":"ID is not a number"}`, http.StatusBadRequest)
		return
	}
	switch {
	case r.Method == http.MethodPost:
		log.Info("API received POST message '/api/task/done'")
		query := `SELECT id, date, repeat FROM scheduler WHERE id = ?`
		err := db.QueryRow(query, taskId).Scan(&task.ID, &task.Date, &task.Repeat)
		if err == sql.ErrNoRows {
			log.Info("SQL there is no task with ID %s (update failed):", taskId)
			http.Error(w, `{"error":"Corresponding ID not found in SQL database"}`, http.StatusNotFound)
			return
		} else if err != nil {
			log.Error("SQL unable to update task with ID %s (%s)", task.ID, err.Error())
			http.Error(w, `{"error":"SQL exec failure"}`, http.StatusInternalServerError)
			return
		}
		log.Debug("API '/api/task/done' retrived from DB %+v", task)
		if task.Repeat == "" {
			log.Debug("SQL 'repeat' is empty, trying to delete task with id: %s", task.ID)
			_, err := db.Exec(`DELETE FROM scheduler WHERE id = ?`, taskId)
			if err != nil {
				log.Error("SQL unable to delete task with ID %s (%s)", task.ID, err.Error())
				http.Error(w, `{"error":"SQL exec failure"}`, http.StatusInternalServerError)
				return
			}
		} else {
			log.Debug("API 'repeat' is not empty, trying to calculate new date for task ID: %s", task.ID)
			now := time.Now()
			nextDate, err := NextDate(now, task.Date, task.Repeat)
			if err != nil {
				log.Debug("API failed to calculate new date for task ID: %s (%s)", task.ID, err.Error())
				http.Error(w, `{"error":"failed to calculate new date for the task"}`, http.StatusInternalServerError)
				return
			}
			log.Debug("API new date calculated, proceed to update DB")
			_, err = db.Exec(`UPDATE scheduler SET date = ? WHERE id = ?`, nextDate, taskId)
			if err != nil {
				log.Error("SQL failed to UPDATE ID: %s (%s)", task.ID, err.Error())
				http.Error(w, `{"error":"SQL exec failure"}`, http.StatusInternalServerError)
				return
			}
			log.Info("SQL task with ID: %s updated", taskId)
		}
		json.NewEncoder(w).Encode(map[string]string{})
	case r.Method == http.MethodDelete:
		log.Info("API received DELETE message '/api/task/done'")
		query := `DELETE FROM scheduler WHERE id = ?`
		execResult, err := db.Exec(query, taskId)
		if err != nil {
			log.Error("SQL unable to delete task with ID %s (%s)", taskId, err.Error())
			http.Error(w, `{"error":"SQL exec failure"}`, http.StatusInternalServerError)
			return
		}
		rowsAffected, err := execResult.RowsAffected()
		if err != nil {
			log.Error("SQL didn't reiceved affected rows from DB after deleting task ID %s (%s)", task.ID, err.Error())
			http.Error(w, `{"error":"Unable to format repeat in PUT message"}`, http.StatusInternalServerError)
			return
		}
		if rowsAffected == 0 {
			log.Info("SQL there is no task with ID (delete failed): %s", task.ID)
			http.Error(w, `{"error":"Corresponding ID not found in SQL database"}`, http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{})
	}
}

func HandlerOTHER(w http.ResponseWriter, r *http.Request) {
	url := r.RequestURI
	method := r.Method
	log.Error("API attempt to access API '%s'(method: %s)", url, method)
	http.Error(w, `{"error": "wrong URL"}`, http.StatusNotFound)
}
