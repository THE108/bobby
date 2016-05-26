package jira

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	dateFormat = "2006-01-02"
	jiraHost   = "jira.lazada.com"
	jiraPath   = "/rest/timesheet-gadget/1.0/raw-timesheet.json"
)

var _ string = `{
  "worklog": [
    {
      "key": "GOPHP-1389",
      "summary": "Add new fields to voucher template",
      "entries": [
        {
          "id": 99920,
          "comment": "",
          "timeSpent": 28800,
          "author": "elmir.khafizov",
          "authorFullName": "Elmir Khafizov",
          "created": 1460384210000,
          "startDate": 1460384160000,
          "updateAuthor": "elmir.khafizov",
          "updateAuthorFullName": "Elmir Khafizov",
          "updated": 1460384210000
        }
      ]
    }
  ],
  "startDate": 1460307600000,
  "endDate": 1460394000000
}`

type Entrie struct {
	ID                   uint64 `json:"id"`
	Comment              string `json:"string"`
	TimeSpent            int64  `json:"timeSpent"`
	Author               string `json:"author"`
	AuthorFullName       string `json:"authorFullName"`
	Created              int64  `json:"created"`
	StartDate            int64  `json:"startDate"`
	UpdateAuthor         string `json:"updateAuthor"`
	UpdateAuthorFullName string `json:"updateAuthorFullName"`
	Updated              int64  `json:"updated"`
}

type WorklogItem struct {
	Key     string   `json:"key"`
	Summary string   `json:"summary"`
	Entries []Entrie `json:"entries"`
}

type Timesheet struct {
	StartDate int64         `json:"startDate"`
	EndDate   int64         `json:"endDate"`
	Worklog   []WorklogItem `json:"worklog"`
}

type UserTimeLog struct {
	Name      string
	TimeSpent time.Duration
}

type Client struct {
	token string
	cli   *http.Client
}

func NewClient(token string) *Client {
	return &Client{
		token: token,
		cli:   &http.Client{},
	}
}

func (this *Client) GetTimesheetForUser(user string, from, to time.Time) (*Timesheet, error) {
	values := url.Values{}
	values.Add("targetUser", user)
	values.Add("startDate", from.Format(dateFormat))
	values.Add("endDate", to.Format(dateFormat))

	jiraURL := url.URL{
		Scheme:   "https",
		Host:     jiraHost,
		Path:     jiraPath,
		RawQuery: values.Encode(),
	}

	//fmt.Printf("url: %s", jiraURL.String())

	req := &http.Request{
		Method:     "GET",
		URL:        &jiraURL,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		// we can use SetBasicAuth here but I think having password in config file is not good idea
		// we'll keep token instead
		Header: make(http.Header),
		Host:   jiraURL.Host,
	}

	req.Header.Set("Authorization", "Basic "+this.token)

	resp, err := this.cli.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var timesheet Timesheet
	if err := json.Unmarshal(responseBody, &timesheet); err != nil {
		return nil, err
	}

	//fmt.Printf("timesheet: %+v\n", timesheet)

	return &timesheet, nil
}

func (this *Client) GetTotalTimeSpentByUser(user string, from, to time.Time) (time.Duration, error) {
	timesheet, err := this.GetTimesheetForUser(user, from, to)
	if err != nil {
		return 0, err
	}

	var totalTimeSpent time.Duration

	// avoid slice elements copying
	for worklogItemIndex := range timesheet.Worklog {
		worklogItem := &timesheet.Worklog[worklogItemIndex]
		for entrieIndex := range worklogItem.Entries {
			entrie := &worklogItem.Entries[entrieIndex]
			spentDuration := time.Duration(entrie.TimeSpent) * time.Second
			totalTimeSpent += spentDuration
			//fmt.Printf("entrie: %+v start: %v created: %v spent: %v\n",
			//	entrie,
			//	time.Unix(entrie.StartDate/1000, 0),
			//	time.Unix(entrie.Created/1000, 0),
			//	spentDuration)
		}
	}

	return totalTimeSpent, nil
}

type durationErrorResult struct {
	totalTimeSpent time.Duration
	user           string
	err            error
}

func (this *Client) getTotalTimeSpentByUserAsync(user string, from, to time.Time, ch chan<- durationErrorResult) {
	totalTimeSpent, err := this.GetTotalTimeSpentByUser(user, from, to)
	//fmt.Printf("Total spent by %s: %v\n", user, totalTimeSpent)

	ch <- durationErrorResult{
		totalTimeSpent: totalTimeSpent,
		user:           user,
		err:            err,
	}
}

func (this *Client) GetUsersLoggedLessThenMin(users map[string]string, from, to time.Time, min time.Duration) ([]UserTimeLog, error) {
	result := make([]UserTimeLog, 0, len(users))
	errors := make([]error, 0, len(users))
	ch := make(chan durationErrorResult)

	for user := range users {
		go this.getTotalTimeSpentByUserAsync(user, from, to, ch)
	}

	for i := 0; i < len(users); i++ {
		res := <-ch

		if res.err != nil {
			errors = append(errors, res.err)
			continue
		}

		if res.totalTimeSpent < min {
			fullName, ok := users[res.user]
			if !ok {
				fullName = res.user
			}

			result = append(result, UserTimeLog{
				Name:      fullName,
				TimeSpent: res.totalTimeSpent,
			})
		}
	}

	if len(errors) > 0 {
		msgs := make([]string, 0, len(errors))
		for _, e := range errors {
			msgs = append(msgs, e.Error())
		}
		return result, fmt.Errorf("Multiple errors occured: %s", strings.Join(msgs, ", "))
	}

	return result, nil
}

func main_jira() {
	token := `YW5kcmV5Y2hlcm5vdjphbGNoRVcxaw==`
	team := map[string]string{
		"andreychernov":    "Andrey Chernov",
		"rustam.zagirov":   "Rustam Zagirov",
		"elmir.khafizov":   "Elmir Khafizov",
		"timur.nurutdinov": "Timur Nurutdinov",
		"evgeny.pak":       "Evgeny Pak",
		"toannguyendinh":   "Toan Nguyen",
	}

	from := time.Date(2016, time.April, 20, 0, 0, 0, 0, time.UTC)
	to := time.Date(2016, time.April, 21, 0, 0, 0, 0, time.UTC)
	minimumTimeSpent := 6 * time.Hour

	c := NewClient(token)

	usersLogs, err := c.GetUsersLoggedLessThenMin(team, from, to, minimumTimeSpent)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	for _, usersLog := range usersLogs {
		fmt.Printf("-> User %s logged only %v from %v to %v\n", usersLog.Name, usersLog.TimeSpent, from, to)
	}
}
