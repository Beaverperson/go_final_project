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
		fmt.Printf("ERROR SQL unable to find working directory\n")
		return nil, err
	}
	dbFile := filepath.Join(appPath, dbFileName)
	fmt.Printf("INFO SQL full path to DB file: %s\n", dbFile)
	_, err = os.Stat(dbFile)
	if err != nil {
		fmt.Print("INFO SQL DB is missing. creating... ")
		os.Create(dbFile)
	}
	dbCreator, errOpen := sql.Open("sqlite3", dbFile)
	if errOpen != nil {
		return nil, fmt.Errorf("unable to open DB \"%s\"", dbFileName)
	}
	fmt.Print("\n")
	_, errCreate := dbCreator.Exec(SQLinit)
	if errCreate != nil {
		return nil, errCreate
	}
	fmt.Printf("INFO SQL DB \"%s\" id ready to use\n", dbFile)
	return dbCreator, nil
}

func NextDate(now time.Time, date string, repeat string) (string, error) {
	dateTime, err := time.Parse(dateFormat, date)
	if err != nil {
		fmt.Printf("ERROR FUNC nextdate incorrect date(%s) format\n", date)
		return "", err
	}
	reDays := regexp.MustCompile(daysRegex)
	switch {
	case repeat == "":
		fmt.Print("DEBUG FUNC nextdate repeat string is missing\n")
		return "", fmt.Errorf("repeat string is missing")
	case repeat == "y":
		dateTime = dateTime.AddDate(1, 0, 0)
		for dateTime.Before(now) || dateTime.Equal(now) {
			dateTime = dateTime.AddDate(1, 0, 0)
		}
		out := dateTime.Format(dateFormat)
		fmt.Printf("DEBUG FUNC nextdate repeat pattern is \"Y\" [%s -> %s]\n", date, out)
		return out, nil
	case strings.HasPrefix(repeat, "d "):
		daysParam := reDays.FindStringSubmatch(repeat)
		if len(daysParam) == 1 {
			fmt.Printf("ERROR FUNC nextdate unable to parce number of days in \"%s\"\n", repeat)
			return "", fmt.Errorf("unable to parce number of days (repeat = %s)", repeat)
		}
		days, _ := strconv.Atoi(daysParam[1])
		if days == 0 || days > 400 {
			fmt.Printf("ERROR FUNC nextdate incorrect days value in %s\n", repeat)
			return "", fmt.Errorf("incorrect number of days %d (repeat = %s)", days, repeat)
		}
		dateTime = dateTime.AddDate(0, 0, days)
		for dateTime.Before(now) || dateTime.Equal(now) {
			dateTime = dateTime.AddDate(0, 0, days)
		}
		out := dateTime.Format(dateFormat)
		fmt.Printf("DEBUG FUNC nextdate repeat pattern is \"D\" [%s -> %s]\n", date, out)
		return out, nil
	default:
		fmt.Printf("ERROR FUNC unable to find correct repeat pattern in \"%s\"\n", repeat)
		return "", fmt.Errorf("unable to execute nextdate lookup")
	}
}
