package timelogs

import (
	"bytes"
	"log"
	"strconv"
	"time"

	"bobby/config"
	"bobby/jira"
	"bobby/utils"
	"fmt"
)

const (
	aproxMessageLength = 128
)

type ISlackClient interface {
	SendMessage(channelID, text string) error
}

type IJiraClient interface {
	GetUsersLoggedLessThenMin([]string, time.Time, time.Time, time.Duration) ([]jira.UserTimeLog, error)
}

type TimelogsDailyMessenger struct {
	Config      *config.Config
	SlackClient ISlackClient
	JiraClient  IJiraClient
}

type userTimeSpentItem struct {
	user      config.User
	timeSpent time.Duration
}

func (this *TimelogsDailyMessenger) Run(now time.Time) {
	usersTimeLogs, err := this.getUsersTimeLogs(now)
	if err != nil {
		log.Printf("Error get users time logs: %s", err.Error())
	}

	log.Printf("usersTimeLogs: %+v", usersTimeLogs)

	jiraLoginToUserMap := make(map[string]config.User, len(this.Config.TimelogsCommand.Team))
	for _, user := range this.Config.TimelogsCommand.Team {
		jiraLoginToUserMap[user.JiraLogin] = user
	}

	userTimeSpentItems := make([]userTimeSpentItem, 0, len(usersTimeLogs))
	for _, item := range usersTimeLogs {
		user, exists := jiraLoginToUserMap[item.Name]
		if !exists {
			continue
		}
		userTimeSpentItems = append(userTimeSpentItems, userTimeSpentItem{
			user:      user,
			timeSpent: item.TimeSpent,
		})
	}

	this.notifyUsers(userTimeSpentItems)

	message := this.render(now, userTimeSpentItems)
	log.Println(message)
	if err := this.SlackClient.SendMessage(this.Config.Slack.Channel, message); err != nil {
		log.Printf("Error send slack message: %s", err.Error())
	}
}

func (this *TimelogsDailyMessenger) getUsersTimeLogs(now time.Time) ([]jira.UserTimeLog, error) {
	log.Printf("start time log\n")
	from, to := utils.GetPreviousDateRange(now)

	usersJiraLogins := make([]string, 0, len(this.Config.TimelogsCommand.Team))
	for _, user := range this.Config.TimelogsCommand.Team {
		usersJiraLogins = append(usersJiraLogins, user.JiraLogin)
	}

	log.Printf("usersJiraLogins: %+v\n", usersJiraLogins)

	return this.JiraClient.GetUsersLoggedLessThenMin(usersJiraLogins, from, to,
		this.Config.TimelogsCommand.MinimumTimeSpent)
}

func (this *TimelogsDailyMessenger) render(now time.Time, userTimeSpentItems []userTimeSpentItem) string {
	var buf bytes.Buffer
	buf.Grow(aproxMessageLength)

	if len(userTimeSpentItems) > 0 {
		rageNumber := 1
		utils.LogIfErr(buf.WriteString("\n :alarm_clock: Time logs:\n"))
		for _, item := range userTimeSpentItems {
			utils.LogIfErr(buf.WriteString("\t "))
			utils.LogIfErr(buf.WriteString(utils.ToSlackUserLogin(item.user.SlackLogin)))
			if item.timeSpent > 0 {
				utils.LogIfErr(buf.WriteString(" logged only "))
				utils.LogIfErr(buf.WriteString(item.timeSpent.String()))
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

func (this *TimelogsDailyMessenger) notifyUsers(userTimeSpentItems []userTimeSpentItem) {
	for _, item := range userTimeSpentItems {
		message := this.renderPersonalMessage(utils.GetFirstName(item.user.Name), item.timeSpent)
		log.Printf("notify user: %s => %s\n", item.user.SlackLogin, message)
		go this.notifyUser(item.user.SlackLogin, message)
	}
}

func (this *TimelogsDailyMessenger) notifyUser(userSlackLogin, message string) {
	if err := this.SlackClient.SendMessage(utils.ToSlackUserLogin(userSlackLogin), message); err != nil {
		log.Printf("send private message error: %s", err.Error())
	}
}

func (this *TimelogsDailyMessenger) renderPersonalMessage(name string, timeSpent time.Duration) string {
	var subMessage string
	if timeSpent == 0 {
		subMessage = "You didn't log any time"
	} else {
		subMessage = fmt.Sprintf("You logged only %v", timeSpent)
	}

	return fmt.Sprintf("Hi, %s! %s for yesterday. Could you please log at least %d hours?",
		name, subMessage, int(this.Config.TimelogsCommand.MinimumTimeSpent.Hours()))
}
