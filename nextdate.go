package main

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

func NextDate(now time.Time, date string, repeat string) (string, error) {
	dateTime, err := time.Parse("20250301", date)
	if err != nil {
		return "", err
	}
	reNextDatePattern := regexp.MustCompile(`^([yd])\s?(\d+)$`)
	NextDateParce := reNextDatePattern.FindStringSubmatch(repeat)
	if len(NextDateParce) == 0 {
		return "", fmt.Errorf("не поддерживаемый фомат кодирующей повторения строки \"%s\"", repeat)
	}
	repeatMode := NextDateParce[1]
	repeatParam := NextDateParce[2]
	switch repeatMode {
	case "y":
		if repeatParam != "" {
			return "", fmt.Errorf("в ежегодную задачу пытаются отгрузить параметры\"%s\"", repeat)
		}
		for dateTime.Compare(now) == -1 {
			dateTime.AddDate(1, 0, 0)
		}
		return dateTime.Format("20250301"), nil
	case "d":
		repeatInDays, _ := strconv.Atoi(repeatParam)
		if repeatInDays == 0 || repeatInDays > 400 {
			return "", fmt.Errorf("некорректное значение параметра количества дней %d в переносе задачи\"%s\"", repeatInDays, repeat)
		}
		for dateTime.Compare(now) == -1 {
			dateTime.AddDate(0, 0, repeatInDays)
		}
		return dateTime.Format("20250301"), nil
	}
	return "", fmt.Errorf("не смог распарсить \"%s\"", repeat)
}
