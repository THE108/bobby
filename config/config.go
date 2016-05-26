package config

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"time"

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
	Pagerduty struct {
		Token     string `yaml:"token"`
		Subdomain string `yaml:"subdomain"`
		Timezone  string `yaml:"timezone"`
	} `yaml:"pagerduty"`
	DutyCommand struct {
		Token       string        `yaml:"token"`
		ScheduleIDs []string      `yaml:"schedule_ids"`
		CacheTTL    time.Duration `yaml:"cache_ttl"`
	} `yaml:"duty_command"`
	TimelogsCommand struct {
		Token            string            `yaml:"token"`
		Team             map[string]string `yaml:"team"`
		MinimumTimeSpent time.Duration     `yaml:"minimum_time_logged"`
		CacheTTL         time.Duration     `yaml:"cache_ttl"`
	} `yaml:"timelogs_command"`
	SendDailyMessageTime string `yaml:"send_daily_message_time"`
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

	if len(cfg.SendDailyMessageTime) == 0 {
		return fmt.Errorf("send daily message time must be non empty")
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

	if len(cfg.Pagerduty.Subdomain) == 0 {
		return fmt.Errorf("pagerduty subdomain must be non empty")
	}

	if len(cfg.Pagerduty.Token) == 0 {
		return fmt.Errorf("pagerduty token must be non empty")
	}

	if len(cfg.DutyCommand.Token) == 0 {
		return fmt.Errorf("duty command token must be non empty")
	}

	if len(cfg.DutyCommand.ScheduleIDs) == 0 {
		return fmt.Errorf("duty command schedule ids must be non empty")
	}

	return nil
}
