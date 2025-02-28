package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func Getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func NextDate(now time.Time, date string, repeat string) (string, error) {
	dateTime, err := time.Parse("20060102", date)
	if err != nil {
		fmt.Printf("ERROR FUNC nextdate incorrect date(%s) format\n", date)
		return "", err
	}
	reDays := regexp.MustCompile(`^d\s(\d+)$`)
	switch {
	case repeat == "":
		fmt.Print("DEBUG FUNC nextdate repeat string is missing\n")
		return "", fmt.Errorf("repeat string is missing")
	case repeat == "y":
		dateTime = dateTime.AddDate(1, 0, 0)
		for dateTime.Before(now) || dateTime.Equal(now) {
			dateTime = dateTime.AddDate(1, 0, 0)
		}
		out := dateTime.Format("20060102")
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
		out := dateTime.Format("20060102")
		fmt.Printf("DEBUG FUNC nextdate repeat pattern is \"Y\" [%s -> %s]\n", date, out)
		return out, nil
	default:
		fmt.Printf("ERROR FUNC unable to find correct repeat pattern in \"%s\"\n", repeat)
		return "", fmt.Errorf("unable to execute nextdate lookup")
	}
}
