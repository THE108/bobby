package timelogs

import (
	"bytes"
	"log"
	"strconv"
	"time"

	"bobby/config"
	"bobby/jira"
	"bobby/utils"
)

const (
	aproxMessageLength = 128
)

type ISlackClient interface {
	SendMessage(channelID, text string) error
}

type IJiraClient interface {
	GetUsersLoggedLessThenMin(map[string]string, time.Time, time.Time, time.Duration) ([]jira.UserTimeLog, error)
}

type TimelogsDailyMessenger struct {
	Config      *config.Config
	SlackClient ISlackClient
	JiraClient  IJiraClient
}

func (this *TimelogsDailyMessenger) Run(now time.Time) {
	usersTimeLogs, err := this.getUsersTimeLogs(now)
	if err != nil {
		log.Printf("Error get users time logs: %s", err.Error())
	}

	message := this.render(now, usersTimeLogs)

	log.Println(message)

	if err := this.SlackClient.SendMessage(this.Config.Slack.Channel, message); err != nil {
		log.Printf("Error send slack message: %s", err)
	}
}

func (this *TimelogsDailyMessenger) getUsersTimeLogs(now time.Time) ([]jira.UserTimeLog, error) {
	log.Printf("start time log\n")
	from, to := utils.GetPreviousDateRange(now)

	userNameToJiraLoginMap := make(map[string]string, len(this.Config.TimelogsCommand.Team))
	for _, user := range this.Config.TimelogsCommand.Team {
		userNameToJiraLoginMap[user.JiraLogin] = user.Name
	}

	log.Printf("userNameToJiraLoginMap: %+v\n", userNameToJiraLoginMap)

	return this.JiraClient.GetUsersLoggedLessThenMin(userNameToJiraLoginMap, from, to,
		this.Config.TimelogsCommand.MinimumTimeSpent)
}

func (this *TimelogsDailyMessenger) render(now time.Time, usersTimeLogs []jira.UserTimeLog) string {
	var buf bytes.Buffer
	buf.Grow(aproxMessageLength)

	if len(usersTimeLogs) > 0 {
		rageNumber := 1
		utils.LogIfErr(buf.WriteString("\n :alarm_clock: Time logs:\n"))
		for _, usersLog := range usersTimeLogs {
			utils.LogIfErr(buf.WriteString("\t "))
			utils.LogIfErr(buf.WriteString(usersLog.Name))
			if usersLog.TimeSpent > 0 {
				utils.LogIfErr(buf.WriteString(" logged only "))
				utils.LogIfErr(buf.WriteString(usersLog.TimeSpent.String()))
				utils.LogIfErr(buf.WriteString("\n"))
			} else {
				utils.LogIfErr(buf.WriteString(" didn't log any time :rage"))
				utils.LogIfErr(buf.WriteString(strconv.Itoa(rageNumber)))
				utils.LogIfErr(buf.WriteString(":\n"))

				rageNumber++
				if rageNumber > 4 {
					rageNumber = 1
				}
			}
		}
	}
	return buf.String()
}
