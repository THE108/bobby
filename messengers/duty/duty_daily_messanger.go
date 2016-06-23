package duty

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"time"

	"bobby/config"
	"bobby/opsgenie"
	"bobby/utils"
)

const (
	timeFormatText = "2006.01.02 15:04"

	aproxMessageLength = 128
)

type ISlackClient interface {
	SendMessage(channelID, text string) error
}

type IDutyProvider interface {
	GetUsersOnDutyForDate(from, to time.Time, scheduleID string) ([]opsgenie.UserOnDuty, error)
}

type DutyDailyMessenger struct {
	Config       *config.Config
	SlackClient  ISlackClient
	DutyProvider IDutyProvider
}

func (this *DutyDailyMessenger) Run(now time.Time) {
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	from, to := dayStart, dayStart.Add(75*time.Hour)
	usersOnDuty, err := this.DutyProvider.GetUsersOnDutyForDate(from, to, this.Config.DutyCommand.ScheduleID)
	if err != nil {
		log.Printf("error get users on duty: %s", err.Error())
		return
	}

	if len(usersOnDuty) == 0 {
		log.Printf("no users on duty found")
		return
	}

	userOnDutyNow, usersOnDutyNext := processUsersOnDuty(now, usersOnDuty)

	this.notifyUsersOnDuty(now, usersOnDutyNext)

	text := this.render(now, userOnDutyNow, usersOnDutyNext)
	log.Printf("text: %s\n", text)

	if err := this.SlackClient.SendMessage(this.Config.Slack.Channel, text); err != nil {
		log.Printf("Error send slack message: %s", err)
	}
}

func processUsersOnDuty(now time.Time, usersOnDuty []opsgenie.UserOnDuty) (opsgenie.UserOnDuty, []opsgenie.UserOnDuty) {
	log.Printf("usersOnDuty: %v\n", usersOnDuty)

	usersOnDutyJoined := opsgenie.JoinDuties(usersOnDuty)

	log.Printf("usersOnDutyJoined: %v\n", usersOnDutyJoined)

	userOnDutyNow := usersOnDutyJoined[0]
	usersOnDutyNext := usersOnDutyJoined[1:]
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

func (this *DutyDailyMessenger) notifyUsersOnDuty(now time.Time, usersOnDuty []opsgenie.UserOnDuty) {
	usersByName := make(map[string]config.User, len(this.Config.TimelogsCommand.Team))
	for _, user := range this.Config.TimelogsCommand.Team {
		usersByName[user.Name] = user
	}

	usersOnDutyByName := opsgenie.JoinDutiesByUserName(usersOnDuty)

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

func renderPrivateMessage(now time.Time, username string, duties []opsgenie.UserOnDuty) string {
	msgs := make([]string, 0, len(duties))
	for _, duty := range duties {
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

func (this *DutyDailyMessenger) render(now time.Time, userOnDutyNow opsgenie.UserOnDuty,
	usersOnDutyNext []opsgenie.UserOnDuty) string {
	var buf bytes.Buffer
	buf.Grow(aproxMessageLength)

	utils.LogIfErr(buf.WriteString(":phone: On duty:\nNow:\n\t"))
	utils.LogIfErr(buf.WriteString(userOnDutyNow.Name))
	utils.LogIfErr(buf.WriteString(" till "))
	utils.LogIfErr(buf.WriteString(userOnDutyNow.End.Format(timeFormatText)))
	utils.LogIfErr(buf.WriteString("\nNext:\n"))

	for _, entrie := range usersOnDutyNext {
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
