package pagerduty

import (
	"log"
	"time"
)

func FilterUsersOnDutyToday(now time.Time, usersOnDuty []UserOnDuty) []UserOnDuty {
	result := make([]UserOnDuty, 0, len(usersOnDuty))
	today := now.Day()
	for _, item := range usersOnDuty {
		if item.End.Before(now) {
			continue
		}
		if item.Start.Day() != today {
			continue
		}
		result = append(result, item)
	}
	return result
}

func JoinDuties(usersOnDuty []UserOnDuty) (usersOnDutyJoined []UserOnDuty) {
	if len(usersOnDuty) == 0 {
		return
	}

	log.Printf("joinDutiesByUserName.usersOnDuty: %v+\n", usersOnDuty)

	// join overlapping intervals
	usersOnDutyJoined = append(make([]UserOnDuty, 0, len(usersOnDuty)), usersOnDuty[0])
	for i := 1; i < len(usersOnDuty); i++ {
		prevIndex := len(usersOnDutyJoined) - 1
		if usersOnDutyJoined[prevIndex].Name == usersOnDuty[i].Name {
			usersOnDutyJoined[prevIndex].End = usersOnDuty[i].End
			continue
		}
		usersOnDutyJoined = append(usersOnDutyJoined, usersOnDuty[i])
	}

	log.Printf("joinDutiesByUserName.usersOnDutyJoined: %+v\n", usersOnDutyJoined)
	return
}

func JoinDutiesByUserName(usersOnDuty []UserOnDuty) map[string][]UserOnDuty {
	// group users on duty by name to avoid notifying users twice
	usersOnDutyByName := make(map[string][]UserOnDuty, len(usersOnDuty))
	for _, userOnDuty := range usersOnDuty {
		usersOnDutyByName[userOnDuty.Name] = append(usersOnDutyByName[userOnDuty.Name], userOnDuty)
	}

	return usersOnDutyByName
}
