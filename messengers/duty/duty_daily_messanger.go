package duty

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"time"

	"bobby/config"
	"bobby/pagerduty"
	"bobby/utils"
)

const (
	timeFormatText = "2006.01.02 15:04"

	aproxMessageLength = 128
)

type ISlackClient interface {
	SendMessage(channelID, text string) error
}

type IPagerDutyClient interface {
	GetUsersOnDutyForDate(from, to time.Time, scheduleIDs ...string) ([]pagerduty.UserOnDuty, error)
}

type DutyDailyMessenger struct {
	Config          *config.Config
	SlackClient     ISlackClient
	PagerdutyClient IPagerDutyClient
}

func (this *DutyDailyMessenger) Run(now time.Time) {
	from, to := now, now.Add(48*time.Hour)
	usersOnDuty, err := this.PagerdutyClient.GetUsersOnDutyForDate(from, to, this.Config.DutyCommand.ScheduleIDs...)
	if err != nil {
		log.Printf("error get users on duty: %s", err.Error())
		return
	}

	if len(usersOnDuty) == 0 {
		log.Printf("no users on duty found")
		return
	}

	usersOnDutyJoined := pagerduty.JoinDuties(usersOnDuty)

	log.Printf("usersOnDuty: %v\n", usersOnDuty)

	this.notifyUsersOnDuty(now, usersOnDutyJoined)

	text := this.render(now, pagerduty.FilterUsersOnDutyToday(now, usersOnDutyJoined))
	if err := this.SlackClient.SendMessage(this.Config.Slack.Channel, text); err != nil {
		log.Printf("Error send slack message: %s", err)
	}
}

func (this *DutyDailyMessenger) notifyUsersOnDuty(now time.Time, usersOnDuty []pagerduty.UserOnDuty) {
	usersByName := make(map[string]config.User, len(this.Config.TimelogsCommand.Team))
	for _, user := range this.Config.TimelogsCommand.Team {
		usersByName[user.Name] = user
	}

	usersOnDutyByName := pagerduty.JoinDutiesByUserName(usersOnDuty)

	log.Printf("notifyUsersOnDuty.usersOnDutyByName: %+v\n", usersOnDutyByName)

	for name, duties := range usersOnDutyByName {
		user, found := usersByName[name]
		if !found {
			log.Printf("can't find user by name: %q", name)
			continue
		}

		message := renderPrivateMessage(now, user.Name, duties)
		log.Printf("message: %q", message)
		if len(message) == 0 {
			continue
		}

		go this.notifyUserOnDuty(user.SlackLogin, message)
	}
}

func renderPrivateMessage(now time.Time, username string, duties []pagerduty.UserOnDuty) string {
	msgs := make([]string, 0, len(duties))
	for _, duty := range duties {
		if duty.End.Before(now) {
			continue
		}
		msgs = append(msgs, fmt.Sprintf("from %s to %s",
			duty.Start.Format(timeFormatText),
			duty.End.Format(timeFormatText)))
	}

	if len(msgs) == 0 {
		return ""
	}

	return fmt.Sprintf("Hello, %s! You are on duty %s. Enjoy!", username, strings.Join(msgs, " and "))
}

func (this *DutyDailyMessenger) notifyUserOnDuty(name, message string) {
	if err := this.SlackClient.SendMessage(utils.ToSlackUserLogin(name), message); err != nil {
		log.Printf("send private message error: %s", err.Error())
	}
}

func (this *DutyDailyMessenger) render(now time.Time, usersOnDuty []pagerduty.UserOnDuty) string {
	var buf bytes.Buffer
	buf.Grow(aproxMessageLength)
	utils.LogIfErr(buf.WriteString(":phone: On duty:\n"))
	for _, entrie := range usersOnDuty {
		utils.LogIfErr(buf.WriteString("\t"))
		utils.LogIfErr(buf.WriteString(entrie.Name))
		utils.LogIfErr(buf.WriteString(" from "))
		utils.LogIfErr(buf.WriteString(entrie.Start.Format(timeFormatText)))
		utils.LogIfErr(buf.WriteString(" to "))
		utils.LogIfErr(buf.WriteString(entrie.End.Format(timeFormatText)))
		utils.LogIfErr(buf.WriteString("\n"))
	}
	return buf.String()
}
