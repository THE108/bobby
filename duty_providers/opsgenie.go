package duty_providers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"
)

const (
	datetimeFormat = "2006-01-02 15:04"
	opsgenieSchema = "https"
	opsgenieHost   = "api.opsgenie.com"
	opsgeniePath   = "/v1/json/schedule/timeline"
)

type scheduleTimeline struct {
	Schedule struct {
		Enabled  bool   `json:"enabled"`
		ID       string `json:"id"`
		Name     string `json:"name"`
		Team     string `json:"team"`
		Timezone string `json:"timezone"`
	} `json:"schedule"`
	Timeline struct {
		StartTime     int64 `json:"startTime"`
		EndTime       int64 `json:"endTime"`
		FinalSchedule struct {
			Rotations []struct {
				ID      string  `json:"id"`
				Name    string  `json:"name"`
				Order   float64 `json:"order"`
				Periods []struct {
					Type       string `json:"type"`
					StartTime  int64  `json:"startTime"`
					EndTime    int64  `json:"endTime"`
					Recipients []struct {
						ID          string `json:"id"`
						Type        string `json:"type"`
						Name        string `json:"name"`
						DisplayName string `json:"displayName"`
					} `json:"recipients"`
				} `json:"periods"`
			} `json:"rotations"`
		} `json:"finalSchedule"`
	} `json:"timeline"`
	Took int `json:"took"`
}

type OpsgenieClient struct {
	apiKey string
}

func NewOpsgenieClient(apiKey string) *OpsgenieClient {
	return &OpsgenieClient{
		apiKey: apiKey,
	}
}

func (this *OpsgenieClient) GetUsersOnDutyForDate(from, to time.Time, scheduleIDs ...string) ([]UserOnDuty, error) {
	if len(scheduleIDs) < 1 {
		return nil, fmt.Errorf("scheduleIDs not provided")
	}

	fmt.Printf("from: %v to: %v\n", from, to)

	interval, err := getDaysInterval(from, to)
	if err != nil {
		return nil, err
	}

	values := url.Values{}
	values.Add("apiKey", this.apiKey)
	values.Add("name", scheduleIDs[0])
	values.Add("interval", strconv.Itoa(interval))
	values.Add("intervalUnit", "days")
	values.Add("date", from.In(time.UTC).Format(datetimeFormat))

	opsgenieURL := url.URL{
		Scheme:   opsgenieSchema,
		Host:     opsgenieHost,
		Path:     opsgeniePath,
		RawQuery: values.Encode(),
	}

	fmt.Printf("url: %s\n", opsgenieURL.String())

	req := &http.Request{
		Method:     "GET",
		URL:        &opsgenieURL,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Host:       opsgenieURL.Host,
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	log.Printf("\n\n%s\n\n", responseBody)

	var timeline scheduleTimeline
	if err := json.Unmarshal(responseBody, &timeline); err != nil {
		return nil, fmt.Errorf("error parse response: %s body: %q", err, responseBody)
	}

	return convertScheduleTimelineToUserOnDuty(&timeline), nil
}

func convertScheduleTimelineToUserOnDuty(timeline *scheduleTimeline) []UserOnDuty {
	usersOnDuty := make([]UserOnDuty, 0, len(timeline.Timeline.FinalSchedule.Rotations))

	for _, rotation := range timeline.Timeline.FinalSchedule.Rotations {
		for _, period := range rotation.Periods {
			if len(period.Recipients) == 0 {
				continue
			}

			usersOnDuty = append(usersOnDuty, UserOnDuty{
				Name:  period.Recipients[0].DisplayName,
				Start: time.Unix(splitSecondsAndNanoseconds(period.StartTime)),
				End:   time.Unix(splitSecondsAndNanoseconds(period.EndTime)),
			})
		}
	}

	log.Printf("usersOnDuty: %v\n", usersOnDuty)

	sort.Sort(ByStartTime(usersOnDuty))

	return usersOnDuty
}

func splitSecondsAndNanoseconds(t int64) (seconds, nanoseconds int64) {
	seconds = t / 1000
	nanoseconds = t - seconds*1000
	return
}

func getDaysInterval(from, to time.Time) (int, error) {
	intervalDuration := to.Sub(from)
	if intervalDuration < 0 {
		return 0, fmt.Errorf("'to' must be after 'from' time period")
	}

	return int(math.Ceil(intervalDuration.Hours() / 24.)), nil
}
