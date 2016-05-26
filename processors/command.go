package processors

import (
	"net/http"
)

// SlackCommand is a struct holding the values that slack will post to our bot
type SlackCommand struct {
	ChannelId   string
	ChannelName string
	UserId      string
	UserName    string
	Command     string
	TeamId      string
	TeamDomain  string
	Text        string
	Token       string
	ResponseURL string
}

// UnmarshalCommand takes the request from slack, returns a SlackCommand object
func UnmarshalCommand(r *http.Request) *SlackCommand {
	return &SlackCommand{
		ChannelId:   r.FormValue("channel_id"),
		ChannelName: r.FormValue("channel_name"),
		UserId:      r.FormValue("user_id"),
		UserName:    r.FormValue("user_name"),
		Command:     r.FormValue("command"),
		TeamId:      r.FormValue("team_id"),
		TeamDomain:  r.FormValue("team_domain"),
		Text:        r.FormValue("text"),
		Token:       r.FormValue("token"),
		ResponseURL: r.FormValue("response_url"),
	}
}
