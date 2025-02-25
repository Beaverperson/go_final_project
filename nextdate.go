package main

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

func NextDate(now time.Time, date string, repeat string) (string, error) {
	dateTime, err := time.Parse("20060102", date)
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
		fmt.Print("DEBUG nextdate function repeat mode is Y\n")
		if repeatParam != "" {
			fmt.Print("DEBUG nextdate function error - в ежегодную задачу пытаются отгрузить параметры\n")
			return "", fmt.Errorf("в ежегодную задачу пытаются отгрузить параметры\"%s\"", repeat)
		}
		for dateTime.Before(now) || dateTime.Equal(now) {
			dateTime = dateTime.AddDate(1, 0, 0)
			fmt.Printf("DEBUG current datetime - %s \n", dateTime.Format("20060102"))
		}
		return dateTime.Format("20060102"), nil
	case "d":
		fmt.Print("DEBUG nextdate function repeat mode is D\n")
		repeatInDays, _ := strconv.Atoi(repeatParam)
		fmt.Printf("DEBUG nextdate function Repeat in days: \"%d\"\n", repeatInDays)
		if repeatInDays == 0 || repeatInDays > 400 {
			fmt.Print("DEBUG nextdate function error - некорректное значение параметра количества дней\n")
			return "", fmt.Errorf("некорректное значение параметра количества дней %d в переносе задачи\"%s\"", repeatInDays, repeat)
		}
		for dateTime.Before(now) || dateTime.Equal(now) {
			dateTime = dateTime.AddDate(0, 0, repeatInDays)
			fmt.Printf("DEBUG current datetime - %s \n", dateTime.Format("20060102"))
		}
		return dateTime.Format("20060102"), nil
	}
	return "", fmt.Errorf("не смог распарсить \"%s\"", repeat)
}
