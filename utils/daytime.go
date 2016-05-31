package utils

import "fmt"

type DayTime struct {
	Hour, Minute int
}

func ParseDayTime(dayTime string) (result DayTime, err error) {
	if len(dayTime) < 5 {
		err = fmt.Errorf("time format error: param length must be 5")
	}
	result.Hour = int((dayTime[0]-'0')*10 + (dayTime[1] - '0'))
	result.Minute = int((dayTime[3]-'0')*10 + (dayTime[4] - '0'))
	if result.Hour < 0 || result.Hour > 23 || result.Minute < 0 || result.Minute > 59 {
		err = fmt.Errorf("time format error: %q", dayTime)
	}
	return
}
