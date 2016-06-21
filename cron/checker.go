package cron

import (
	"time"

	"bobby/utils"
)

type everyWorkingDayAtTimeChecker struct {
	dayTime utils.DayTime
}

func EveryWorkingDayAt(dayTime utils.DayTime) IChecker {
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
