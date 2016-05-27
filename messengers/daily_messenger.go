package messengers

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"bobby/config"
	"bobby/jira"
	"bobby/pagerduty"
	"bobby/utils"
)

const (
	dateFormatText = "02 Jan, Mon"
	timeFormatText = "2006.01.02 15:04"

	aproxMessageLength = 128
)

type ISlackClient interface {
	SendMessage(channelID, text string) error
}

type IPagerDutyClient interface {
	GetUsersOnDutyForDate(from, to time.Time, scheduleIDs ...string) ([]pagerduty.UserOnDuty, error)
}

type IJiraClient interface {
	GetUsersLoggedLessThenMin(map[string]string, time.Time, time.Time, time.Duration) ([]jira.UserTimeLog, error)
}

type DailyMessenger struct {
	Config          *config.Config
	SlackClient     ISlackClient
	PagerdutyClient IPagerDutyClient
	JiraClient      IJiraClient
}

func (this *DailyMessenger) Run() {
	now := time.Now()
	resultChan := make(chan interface{})

	go this.getUsersOnDuty(now, resultChan)
	go this.getUsersTimeLogs(now, resultChan)

	var usersOnDuty []pagerduty.UserOnDuty
	var usersTimeLogs []jira.UserTimeLog
	var text bytes.Buffer
	for i := 0; i < 2; i++ {
		result := <-resultChan
		switch res := result.(type) {
		case error:
			log.Printf("Error: %s", res.Error())
			logIfErr(text.WriteString(res.Error()))
			logIfErr(text.WriteString("\n"))

		case []pagerduty.UserOnDuty:
			usersOnDuty = res

		case []jira.UserTimeLog:
			usersTimeLogs = res
		}
	}

	usersOnDutyJoined := joinDuties(usersOnDuty)

	log.Printf("usersOnDuty: %v\n", usersOnDuty)

	this.notifyUsersOnDuty(now, usersOnDutyJoined)

	this.render(&text, filterUsersOnDutyToday(now, usersOnDutyJoined), usersTimeLogs, now)

	log.Println(text.String())

	if err := this.SlackClient.SendMessage(this.Config.Slack.Channel, text.String()); err != nil {
		log.Printf("Error send slack message: %s", err)
	}
}

func filterUsersOnDutyToday(now time.Time, usersOnDuty []pagerduty.UserOnDuty) []pagerduty.UserOnDuty {
	result := make([]pagerduty.UserOnDuty, 0, len(usersOnDuty))
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

func joinDuties(usersOnDuty []pagerduty.UserOnDuty) (usersOnDutyJoined []pagerduty.UserOnDuty) {
	if len(usersOnDuty) == 0 {
		return
	}

	log.Printf("joinDutiesByUserName.usersOnDuty: %v+\n", usersOnDuty)

	// join overlapping intervals
	usersOnDutyJoined = append(make([]pagerduty.UserOnDuty, 0, len(usersOnDuty)), usersOnDuty[0])
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

func joinDutiesByUserName(usersOnDuty []pagerduty.UserOnDuty) map[string][]pagerduty.UserOnDuty {
	// group users on duty by name to avoid notifying users twice
	usersOnDutyByName := make(map[string][]pagerduty.UserOnDuty, len(usersOnDuty))
	for _, userOnDuty := range usersOnDuty {
		usersOnDutyByName[userOnDuty.Name] = append(usersOnDutyByName[userOnDuty.Name], userOnDuty)
	}

	return usersOnDutyByName
}

func (this *DailyMessenger) notifyUsersOnDuty(now time.Time, usersOnDuty []pagerduty.UserOnDuty) {
	usersByName := make(map[string]config.User, len(this.Config.TimelogsCommand.Team))
	for _, user := range this.Config.TimelogsCommand.Team {
		usersByName[user.Name] = user
	}

	usersOnDutyByName := joinDutiesByUserName(usersOnDuty)

	log.Printf("notifyUsersOnDuty.usersOnDutyByName: %+v\n", usersOnDutyByName)

	for name, duties := range usersOnDutyByName {
		msgs := make([]string, 0, len(duties))
		for _, duty := range duties {
			if duty.End.Before(now) {
				continue
			}
			msgs = append(msgs, fmt.Sprintf("from %s to %s",
				duty.Start.Format(timeFormatText),
				duty.End.Format(timeFormatText)))
		}

		user, found := usersByName[name]
		if !found {
			continue
		}

		if len(msgs) == 0 {
			continue
		}

		message := fmt.Sprintf("Hello, %s! You are on duty %s. Enjoy!", user.Name, strings.Join(msgs, " and "))
		log.Printf("message: %q", message)
		if err := this.SlackClient.SendMessage(toSlackUserLogin(user.SlackLogin), message); err != nil {
			log.Printf("send private message error: %s", err.Error())
		}
	}
}

func toSlackUserLogin(name string) string {
	if strings.HasPrefix(name, "@") {
		return name
	}
	return "@" + name
}

func (this *DailyMessenger) getUsersTimeLogs(now time.Time, resultChan chan<- interface{}) {
	log.Printf("start time log\n")
	from, to := utils.GetPreviousDateRange(now)

	userNameToJiraLoginMap := make(map[string]string, len(this.Config.TimelogsCommand.Team))
	for _, user := range this.Config.TimelogsCommand.Team {
		userNameToJiraLoginMap[user.JiraLogin] = user.Name
	}

	log.Printf("userNameToJiraLoginMap: %+v\n", userNameToJiraLoginMap)

	usersLogs, err := this.JiraClient.GetUsersLoggedLessThenMin(userNameToJiraLoginMap, from, to,
		this.Config.TimelogsCommand.MinimumTimeSpent)
	log.Printf("time log\n")
	if err != nil {
		resultChan <- err
		return
	}

	log.Printf("usersLogs: %+v\n", usersLogs)

	resultChan <- usersLogs
}

func (this *DailyMessenger) getUsersOnDuty(now time.Time, resultChan chan<- interface{}) {
	log.Printf("start duty\n")
	from, to := now, now.Add(48*time.Hour)
	result, err := this.PagerdutyClient.GetUsersOnDutyForDate(from, to, this.Config.DutyCommand.ScheduleIDs...)
	log.Printf("duty\n")
	if err != nil {
		resultChan <- err
		return
	}
	resultChan <- result
}

func (this *DailyMessenger) render(buf *bytes.Buffer, usersOnDuty []pagerduty.UserOnDuty, usersTimeLogs []jira.UserTimeLog,
	now time.Time) {
	buf.Grow(aproxMessageLength)
	logIfErr(buf.WriteString("Good morning, bobbers!\nToday is "))
	logIfErr(buf.WriteString(now.Format(dateFormatText)))

	if len(usersOnDuty) > 0 {
		logIfErr(buf.WriteString("\n :phone: On duty:\n"))
		for _, entrie := range usersOnDuty {
			logIfErr(buf.WriteString("\t"))
			logIfErr(buf.WriteString(entrie.Name))
			logIfErr(buf.WriteString(" from "))
			logIfErr(buf.WriteString(entrie.Start.Format(timeFormatText)))
			logIfErr(buf.WriteString(" to "))
			logIfErr(buf.WriteString(entrie.End.Format(timeFormatText)))
			logIfErr(buf.WriteString("\n"))
		}
	}

	if len(usersTimeLogs) > 0 {
		rageNumber := 1
		logIfErr(buf.WriteString("\n :alarm_clock: Time logs:\n"))
		for _, usersLog := range usersTimeLogs {
			logIfErr(buf.WriteString("\t "))
			logIfErr(buf.WriteString(usersLog.Name))
			if usersLog.TimeSpent > 0 {
				logIfErr(buf.WriteString(" logged only "))
				logIfErr(buf.WriteString(usersLog.TimeSpent.String()))
				logIfErr(buf.WriteString("\n"))
			} else {
				logIfErr(buf.WriteString(" didn't log any time :rage"))
				logIfErr(buf.WriteString(strconv.Itoa(rageNumber)))
				logIfErr(buf.WriteString(":\n"))

				rageNumber++
				if rageNumber > 4 {
					rageNumber = 1
				}
			}
		}
	}
}

func logIfErr(_ int, err error) {
	if err != nil {
		log.Println(err.Error())
	}
}
