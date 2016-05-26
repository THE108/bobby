package processors

import (
	"fmt"
	"time"

	"bobby/pagerduty"
	"bobby/utils"
	"strings"
)

type IPagerDutyClient interface {
	GetUsersOnDutyForDate(from, to time.Time, scheduleIDs ...string) ([]pagerduty.UserOnDuty, error)
}

type DutyCommandProcessor struct {
	PagerdutyClient IPagerDutyClient
	ScheduleIDs     []string
	from, to        time.Time
}

func (this *DutyCommandProcessor) Init(args []string, now time.Time) (err error) {
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
	// if args[0] is a user name in list

	usersOnDuty, err := this.PagerdutyClient.GetUsersOnDutyForDate(this.from, this.to, this.ScheduleIDs...)
	if err != nil {
		return "", err
	}

	return this.renderText(usersOnDuty), nil
}

func (this *DutyCommandProcessor) renderText(usersOnDuty []pagerduty.UserOnDuty) string {
	text := fmt.Sprintf("On duty %s:\n", this.from.Format(dateFormatText))
	for _, item := range usersOnDuty {
		text += fmt.Sprintf("\t%s from %s to %s\n",
			item.Name,
			item.Start.Format(timeFormatText),
			item.End.Format(timeFormatText))
	}
	return text
}
