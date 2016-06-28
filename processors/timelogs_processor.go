package processors

import (
	"fmt"
	"strings"
	"time"

	"bobby/jira"
	"bobby/utils"
)

type IJiraClient interface {
	GetUsersLoggedLessThenMin([]string, time.Time, time.Time, time.Duration) ([]jira.UserTimeLog, error)
}

type TimeLogsCommandProcessor struct {
	JiraClient       IJiraClient
	Users            []string
	MinimumTimeSpent time.Duration
	from, to         time.Time
}

func (this *TimeLogsCommandProcessor) Init(args []string, now time.Time) error {
	if len(args) == 0 {
		this.from, this.to = utils.GetPreviousDateRange(now)
		return nil
	}

	date, err := utils.GetDateFromArgs(args[0], now)
	if err != nil {
		this.from, this.to = utils.GetPreviousDateRange(now)
		return err
	}

	this.from, this.to = utils.GetPreviousDateRange(date)
	return nil
}

func (this *TimeLogsCommandProcessor) GetCacheKey() string {
	return strings.Join([]string{this.from.Format(dateFormatText), this.to.Format(dateFormatText)}, "_")
}

func (this *TimeLogsCommandProcessor) Process() (string, error) {
	usersLogs, err := this.JiraClient.GetUsersLoggedLessThenMin(this.Users, this.from, this.to, this.MinimumTimeSpent)
	if err != nil {
		return "", err
	}
	return this.renderText(usersLogs), nil
}

func (this *TimeLogsCommandProcessor) renderText(usersTimeLogs []jira.UserTimeLog) string {
	var text string
	if len(usersTimeLogs) > 0 {
		rageNumber := 1
		for _, usersLog := range usersTimeLogs {
			text += usersLog.Name
			if usersLog.TimeSpent > 0 {
				text += fmt.Sprintf(" logged only %v\n", usersLog.TimeSpent)
			} else {
				text += fmt.Sprintf(" didn't log any time :rage%d:\n", rageNumber)
				rageNumber++
				if rageNumber > 4 {
					rageNumber = 1
				}
			}
		}
	} else {
		text = fmt.Sprintf("\n :simple_smile: No users with logged time less then %v\n", this.MinimumTimeSpent)
	}
	return text
}
