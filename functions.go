package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "main/logging"

	_ "github.com/mattn/go-sqlite3"
)

func Getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func GetDBConnector(dbFileName string) (*sql.DB, error) {
	appPath, err := os.Getwd()
	if err != nil {
		log.Error("SQL unable to find working directory (%s)", err.Error())
		return nil, err
	}
	dbFile := filepath.Join(appPath, dbFileName)
	log.Info("SQL full path to DB file: %s", dbFile)
	_, err = os.Stat(dbFile)
	if err != nil {
		log.Info("SQL DB is missing. creating... ")
		os.Create(dbFile)
	}
	dbCreator, errOpen := sql.Open("sqlite3", dbFile)
	if errOpen != nil {
		log.Error("SQL unable to open '%s' working directory (%s)", dbFile, err.Error())
		return nil, fmt.Errorf("unable to open DB '%s'", dbFileName)
	}
	_, errCreate := dbCreator.Exec(SQLinit)
	if errCreate != nil {
		return nil, errCreate
	}
	log.Info("SQL DB '%s' id ready to use", dbFile)
	return dbCreator, nil
}

func NextDate(now time.Time, date string, repeat string) (string, error) {
	dateTime, err := time.Parse(dateFormat, date)
	if err != nil {
		log.Error("FUNC nextdate incorrect date '%s' format (%s)", date, err.Error())
		return "", err
	}
	reDays := regexp.MustCompile(daysRegex)
	switch {
	case repeat == "":
		log.Debug("FUNC nextdate repeat string is missing")
		return "", fmt.Errorf("repeat string is missing")
	case repeat == "y":
		dateTime = dateTime.AddDate(1, 0, 0)
		for dateTime.Before(now) || dateTime.Equal(now) {
			dateTime = dateTime.AddDate(1, 0, 0)
		}
		out := dateTime.Format(dateFormat)
		log.Debug("FUNC nextdate repeat pattern is 'Y' [%s -> %s]", date, out)
		return out, nil
	case strings.HasPrefix(repeat, "d "):
		daysParam := reDays.FindStringSubmatch(repeat)
		if len(daysParam) == 1 {
			log.Error("FUNC nextdate unable to parce number of days in '%s'", repeat)
			return "", fmt.Errorf("unable to parce number of days (repeat = %s)", repeat)
		}
		days, _ := strconv.Atoi(daysParam[1])
		if days == 0 || days > 400 {
			log.Error("FUNC nextdate incorrect days value in %s", repeat)
			return "", fmt.Errorf("incorrect number of days %d (repeat = %s)", days, repeat)
		}
		dateTime = dateTime.AddDate(0, 0, days)
		for dateTime.Before(now) || dateTime.Equal(now) {
			dateTime = dateTime.AddDate(0, 0, days)
		}
		out := dateTime.Format(dateFormat)
		log.Debug("FUNC nextdate repeat pattern is 'D' [%s -> %s]", date, out)
		return out, nil
	default:
		log.Error("FUNC unable to find correct repeat pattern in '%s'", repeat)
		return "", fmt.Errorf("unable to execute nextdate lookup")
	}
}
