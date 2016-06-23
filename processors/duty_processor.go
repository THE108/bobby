package processors

import (
	"bytes"
	"strings"
	"time"

	"bobby/opsgenie"
	"bobby/utils"
)

const (
	aproxMessageLength = 128
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

	currentUserOnDuty, nextUsersOnDuty := opsgenie.SplitCurrentAndNextUsersOnDuty(this.now, usersOnDuty)

	return this.renderText(currentUserOnDuty, nextUsersOnDuty), nil
}

func (this *DutyCommandProcessor) renderText(userOnDutyNow opsgenie.UserOnDuty, usersOnDuty []opsgenie.UserOnDuty) string {
	var buf bytes.Buffer
	buf.Grow(aproxMessageLength)
	utils.LogIfErr(buf.WriteString(":phone: On duty:\nNow:\n\t"))
	utils.LogIfErr(buf.WriteString(userOnDutyNow.Name))
	utils.LogIfErr(buf.WriteString(" till "))
	utils.LogIfErr(buf.WriteString(userOnDutyNow.End.Format(timeFormatText)))
	utils.LogIfErr(buf.WriteString("\nNext:\n"))

	for _, item := range usersOnDuty {
		utils.LogIfErr(buf.WriteString("\t"))
		utils.LogIfErr(buf.WriteString(item.Name))
		utils.LogIfErr(buf.WriteString(" from "))
		utils.LogIfErr(buf.WriteString(item.Start.Format(timeFormatText)))
		utils.LogIfErr(buf.WriteString(" to "))
		utils.LogIfErr(buf.WriteString(item.End.Format(timeFormatText)))
		utils.LogIfErr(buf.WriteString("\n"))
	}
	return buf.String()
}
