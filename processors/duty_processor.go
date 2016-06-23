package processors

import (
	"fmt"
	"time"

	"bobby/opsgenie"
	"bobby/utils"
	"strings"
)

type IDutyProvider interface {
	GetUsersOnDutyForDate(from, to time.Time, scheduleID string) ([]opsgenie.UserOnDuty, error)
}

type DutyCommandProcessor struct {
	DutyProvider  IDutyProvider
	ScheduleID    string
	from, to, now time.Time
}

func (this *DutyCommandProcessor) Init(args []string, now time.Time) (err error) {
	this.now = now
	if len(args) == 0 {
		this.from = now
		this.to = now.Add(24 * time.Hour)
		return
	}
	this.from, err = utils.GetDateFromArgs(args[0], now)
	this.to = this.from.Add(24 * time.Hour)
	return
}

func (this *DutyCommandProcessor) GetCacheKey() string {
	return strings.Join([]string{this.from.Format(dateFormatText), this.to.Format(dateFormatText)}, "_")
}

func (this *DutyCommandProcessor) Process() (string, error) {
	usersOnDuty, err := this.DutyProvider.GetUsersOnDutyForDate(this.from, this.to, this.ScheduleID)
	if err != nil {
		return "", err
	}

	usersOnDuty = opsgenie.FilterUsersOnDutyToday(this.now, opsgenie.JoinDuties(usersOnDuty))

	return this.renderText(usersOnDuty), nil
}

func (this *DutyCommandProcessor) renderText(usersOnDuty []opsgenie.UserOnDuty) string {
	text := fmt.Sprintf("On duty %s:\n", this.from.Format(dateFormatText))
	for _, item := range usersOnDuty {
		text += fmt.Sprintf("\t%s from %s to %s\n",
			item.Name,
			item.Start.Format(timeFormatText),
			item.End.Format(timeFormatText))
	}
	return text
}
