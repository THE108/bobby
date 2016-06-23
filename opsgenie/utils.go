package opsgenie

import (
	"time"
)

const (
	dateFormat = "2006-01-02"
)

type UserOnDuty struct {
	Name       string
	Start, End time.Time
}

type ByStartTime []UserOnDuty

func (a ByStartTime) Len() int           { return len(a) }
func (a ByStartTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByStartTime) Less(i, j int) bool { return a[i].Start.Before(a[j].Start) }

func FilterUsersOnDutyToday(now time.Time, usersOnDuty []UserOnDuty) []UserOnDuty {
	result := make([]UserOnDuty, 0, len(usersOnDuty))
	today := now.Day()
	tomorrow := now.Add(24 * time.Hour).Day()
	for _, item := range usersOnDuty {
		if item.End.Before(now) {
			continue
		}
		day := item.Start.Day()
		if day != today && day != tomorrow {
			continue
		}
		result = append(result, item)
	}
	return result
}

func FilterUsersOnDutyByDate(now, limit time.Time, usersOnDuty []UserOnDuty) []UserOnDuty {
	result := make([]UserOnDuty, 0, len(usersOnDuty))
	for _, item := range usersOnDuty {
		if item.End.Before(now) || item.Start.After(limit) {
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

	//log.Printf("joinDutiesByUserName.usersOnDuty: %v+\n", usersOnDuty)

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

	//log.Printf("joinDutiesByUserName.usersOnDutyJoined: %+v\n", usersOnDutyJoined)
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

func SplitCurrentAndNextUsersOnDuty(now time.Time, usersOnDuty []UserOnDuty) (UserOnDuty, []UserOnDuty) {
	if len(usersOnDuty) == 0 {
		return UserOnDuty{}, nil
	}

	userOnDutyNow := usersOnDuty[0]
	usersOnDutyNext := usersOnDuty[1:]
	for {
		if userOnDutyNow.Start.Before(now) && userOnDutyNow.End.After(now) {
			break
		}

		if len(usersOnDutyNext) == 0 {
			break
		}

		userOnDutyNow = usersOnDutyNext[0]
		usersOnDutyNext = usersOnDutyNext[1:]
	}
	return userOnDutyNow, usersOnDutyNext
}
