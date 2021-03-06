package jira

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/codeship/go-retro"
)

const (
	dateFormat = "2006-01-02"
	jiraHost   = "jira.lazada.com"
	jiraPath   = "/rest/timesheet-gadget/1.0/raw-timesheet.json"

	maxRetryAttempts = 3
)

var emptyResponseError = fmt.Errorf("empty response")

func isEmptyResponseError(e error) bool {
	return e.Error() == emptyResponseError.Error()
}

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

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	log.Printf("JIRA response for %q from %v to %v: %q", user, from, to, responseBody)

	var timesheet Timesheet
	if err := json.Unmarshal(responseBody, &timesheet); err != nil {
		return nil, err
	}

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
			if entrie.Author != user {
				return 0, fmt.Errorf("worklog author %q != user %q", entrie.Author, user)
			}

			spentDuration := time.Duration(entrie.TimeSpent) * time.Second
			totalTimeSpent += spentDuration
			//log.Printf("entrie: %+v start: %v created: %v spent: %v\n",
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
	var totalTimeSpent time.Duration
	getTimesheetError := retro.DoWithRetry(func() error {
		result, err := this.GetTotalTimeSpentByUser(user, from, to)
		if err != nil {
			return retro.NewBackoffRetryableError(err, maxRetryAttempts)
		}

		if result == 0 {
			return retro.NewBackoffRetryableError(emptyResponseError, maxRetryAttempts)
		}

		totalTimeSpent = result
		return nil
	})

	ch <- durationErrorResult{
		totalTimeSpent: totalTimeSpent,
		user:           user,
		err:            getTimesheetError,
	}
}

func (this *Client) GetUsersLoggedLessThenMin(users []string, from, to time.Time, min time.Duration) ([]UserTimeLog, error) {
	result := make([]UserTimeLog, 0, len(users))
	errors := make([]error, 0, len(users))
	ch := make(chan durationErrorResult)

	for _, user := range users {
		go this.getTotalTimeSpentByUserAsync(user, from, to, ch)
	}

	for i := 0; i < len(users); i++ {
		res := <-ch

		if res.err != nil && !isEmptyResponseError(res.err) {
			errors = append(errors, res.err)
			continue
		}

		if res.totalTimeSpent < min {
			result = append(result, UserTimeLog{
				Name:      res.user,
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
