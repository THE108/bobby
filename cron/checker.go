package cron

import (
	"fmt"
	"time"
)

type DayTime struct {
	Hour, Minute int
}

func ParseDayTime(dayTime string) (result DayTime, err error) {
	result.Hour = int((dayTime[0]-'0')*10 + (dayTime[1] - '0'))
	result.Minute = int((dayTime[3]-'0')*10 + (dayTime[4] - '0'))
	if result.Hour < 0 || result.Hour > 23 || result.Minute < 0 || result.Minute > 59 {
		err = fmt.Errorf("time format error: %q", dayTime)
	}
	return
}

type everyWorkingDayAtTimeChecker struct {
	dayTime DayTime
}

func EveryWorkingDayAt(dayTime DayTime) IChecker {
	return &everyWorkingDayAtTimeChecker{
		dayTime: dayTime,
	}
}

func (this *everyWorkingDayAtTimeChecker) Check(now time.Time) bool {
	//switch now.Weekday() {
	//case time.Sunday, time.Saturday:
	//	return false
	//}

	hour, minute, _ := now.Clock()
	if hour != this.dayTime.Hour || minute != this.dayTime.Minute {
		return false
	}

	return true
}
