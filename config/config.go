package config

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"time"

	"bobby/utils"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Main struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`
	} `yaml:"main"`
	Slack struct {
		Token   string `yaml:"token"`
		Channel string `yaml:"channel"`
	} `yaml:"slack"`
	Jira struct {
		Token string `yaml:"token"`
	} `yaml:"jira"`
	Opsgenie struct {
		Token string `yaml:"token"`
	} `yaml:"opsgenie"`
	DutyCommand struct {
		Enable                 bool          `yaml:"enable"`
		Name                   string        `yaml:"name"`
		Token                  string        `yaml:"token"`
		ScheduleID             string        `yaml:"schedule-id"`
		CacheTTL               time.Duration `yaml:"cache-ttl"`
		DailyMessageTimeString string        `yaml:"daily-message-time"`
		DailyMessageTime       utils.DayTime `yaml:"-"`
	} `yaml:"duty-command"`
	TimelogsCommand struct {
		Enable                 bool          `yaml:"enable"`
		Name                   string        `yaml:"name"`
		Token                  string        `yaml:"token"`
		Team                   []User        `yaml:"team"`
		MinimumTimeSpent       time.Duration `yaml:"minimum-time-logged"`
		CacheTTL               time.Duration `yaml:"cache-ttl"`
		DailyMessageTimeString string        `yaml:"daily-message-time"`
		DailyMessageTime       utils.DayTime `yaml:"-"`
	} `yaml:"timelogs-command"`
}

func ParseConfig(filename string) (*Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if err := validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func validate(cfg *Config) error {
	port, err := strconv.Atoi(cfg.Main.Port)
	if err != nil {
		return err
	}

	if port <= 0 || port >= 60000 {
		return fmt.Errorf("port value must be positive and less then 60000")
	}

	if len(cfg.Slack.Token) == 0 {
		return fmt.Errorf("slack token must be non empty")
	}

	if len(cfg.Slack.Channel) == 0 {
		return fmt.Errorf("slack channel must be non empty")
	}

	if len(cfg.Jira.Token) == 0 {
		return fmt.Errorf("jira token must be non empty")
	}

	if len(cfg.DutyCommand.DailyMessageTimeString) == 0 {
		return fmt.Errorf("duty daily message time must be non empty")
	}

	if len(cfg.DutyCommand.Name) == 0 {
		return fmt.Errorf("empty duty command name")
	}

	if len(cfg.Opsgenie.Token) == 0 {
		return fmt.Errorf("opsgenie token must be non empty")
	}

	dutyDailyMessageTime, err := utils.ParseDayTime(cfg.DutyCommand.DailyMessageTimeString)
	if err != nil {
		return fmt.Errorf("error parse duty daily message date time: %s", err.Error())
	}
	cfg.DutyCommand.DailyMessageTime = dutyDailyMessageTime

	timlogsDailyMessageTime, err := utils.ParseDayTime(cfg.TimelogsCommand.DailyMessageTimeString)
	if err != nil {
		return fmt.Errorf("error parse timelogs daily message date time: %s", err.Error())
	}
	cfg.TimelogsCommand.DailyMessageTime = timlogsDailyMessageTime

	if len(cfg.TimelogsCommand.Name) == 0 {
		return fmt.Errorf("empty time logs command name")
	}

	if len(cfg.TimelogsCommand.Token) == 0 {
		return fmt.Errorf("timelogs command auth token must be non empty")
	}

	if cfg.TimelogsCommand.MinimumTimeSpent == 0 {
		return fmt.Errorf("timelogs command minimum time logged")
	}

	if len(cfg.TimelogsCommand.Team) == 0 {
		return fmt.Errorf("timelogs command team must be non empty")
	}

	if len(cfg.DutyCommand.Token) == 0 {
		return fmt.Errorf("duty command token must be non empty")
	}

	if len(cfg.DutyCommand.ScheduleID) == 0 {
		return fmt.Errorf("duty command schedule id must be non empty")
	}

	return nil
}
