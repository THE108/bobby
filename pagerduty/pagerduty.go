package pagerduty

import (
	"fmt"
	"sort"
	"time"

	pgclient "github.com/danryan/go-pagerduty/pagerduty"
)

const (
	dateFormat = "2006-01-02"
)

type UserOnDuty struct {
	Name       string
	Start, End time.Time
}

type ByStartTime []UserOnDuty

func (a ByStartTime) Len() int           { return len(a) }
func (a ByStartTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByStartTime) Less(i, j int) bool { return a[i].Start.Before(a[j].Start) }

type Client struct {
	timezone string
	pd       *pgclient.Client
}

func NewClient(subdomain, apiKey, timezone string) *Client {
	return &Client{
		pd:       pgclient.New(subdomain, apiKey),
		timezone: timezone,
	}
}

func (this *Client) getUsersOnDutyForDate(from, to time.Time, scheduleID string) ([]UserOnDuty, error) {
	opt := &pgclient.ScheduleEntriesOptions{
		Since:    from.Format(dateFormat),
		Until:    to.Format(dateFormat),
		Timezone: this.timezone,
	}

	entries, resp, err := this.pd.Schedules.Entries(scheduleID, opt)
	if err != nil {
		return nil, fmt.Errorf("Error: %s resp: %+v\n", err, resp)
	}

	result := make([]UserOnDuty, len(entries.ScheduleEntries))
	for i, entrie := range entries.ScheduleEntries {
		result[i].Name = entrie.User.Name
		start, err := time.Parse(time.RFC3339, entrie.Start)
		if err != nil {
			return nil, fmt.Errorf("error parse start time: %s", err)
		}
		end, err := time.Parse(time.RFC3339, entrie.End)
		if err != nil {
			return nil, fmt.Errorf("error parse end time: %s", err)
		}
		result[i].Start = start
		result[i].End = end
	}
	return result, nil
}

func (this *Client) getUsersOnDutyForDateAsync(ch chan<- interface{}, from, to time.Time, scheduleID string) {
	entries, err := this.getUsersOnDutyForDate(from, to, scheduleID)
	if err != nil {
		ch <- err
		return
	}
	ch <- entries
}

func (this *Client) GetUsersOnDutyForDate(from, to time.Time, scheduleIDs ...string) ([]UserOnDuty, error) {
	ch := make(chan interface{})
	for _, scheduleID := range scheduleIDs {
		go this.getUsersOnDutyForDateAsync(ch, from, to, scheduleID)
	}

	// preallocate allEntries slice, assume there are 10 entries in duty schedule
	allEntries := make([]UserOnDuty, 0, 10)
	var err error
	for range scheduleIDs {
		optional := <-ch
		switch result := optional.(type) {
		case []UserOnDuty:
			allEntries = append(allEntries, result...)
		case error:
			err = result
		}
	}

	if err != nil {
		return nil, err
	}

	sort.Sort(ByStartTime(allEntries))
	return allEntries, nil
}
