package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"

	"bobby/cache"
	"bobby/config"
	"bobby/cron"
	"bobby/jira"
	"bobby/messengers/duty"
	"bobby/messengers/timelogs"
	"bobby/opsgenie"
	"bobby/processors"
	"bobby/slack"
)

const (
	DefaultCacheSize = 256
)

func initCommandProcessManager(cfg *config.Config, slackClient processors.ISlackPostponedClient, cache processors.ICache,
	dutyProvider processors.IDutyProvider, jiraClient processors.IJiraClient) *processors.CommandProcessManager {
	commandProcessManager := processors.NewCommandProcessManager()
	commandProcessManager.AddCommandProcessor(cfg.DutyCommand.Name, &processors.PostponedCommandProcessor{
		Token:         cfg.DutyCommand.Token,
		SlackClient:   slackClient,
		Cache:         cache,
		CacheDuration: cfg.DutyCommand.CacheTTL,
		Processor: &processors.DutyCommandProcessor{
			DutyProvider: dutyProvider,
			ScheduleID:   cfg.DutyCommand.ScheduleID,
		},
	})

	usersJiraLogins := make([]string, 0, len(cfg.TimelogsCommand.Team))
	for _, user := range cfg.TimelogsCommand.Team {
		usersJiraLogins = append(usersJiraLogins, user.JiraLogin)
	}

	commandProcessManager.AddCommandProcessor(cfg.TimelogsCommand.Name, &processors.PostponedCommandProcessor{
		Token:         cfg.TimelogsCommand.Token,
		SlackClient:   slackClient,
		Cache:         cache,
		CacheDuration: cfg.TimelogsCommand.CacheTTL,
		Processor: &processors.TimeLogsCommandProcessor{
			JiraClient:       jiraClient,
			Users:            usersJiraLogins,
			MinimumTimeSpent: cfg.TimelogsCommand.MinimumTimeSpent,
		},
	})
	return commandProcessManager
}

func initHandlers(mux *http.ServeMux, commandProcessManager *processors.CommandProcessManager) {
	mux.HandleFunc("/api/v1", func(w http.ResponseWriter, r *http.Request) {
		command := processors.UnmarshalCommand(r)
		log.Printf("command: %+v\n", command)
		result, err := commandProcessManager.ProcessCommand(command)
		if err != nil {
			fmt.Fprintf(w, "Error: %q", err.Error())
			return
		}

		if !result.Postponed {
			fmt.Fprint(w, result.Text)
		}
	})
}

func run(addr string, mux *http.ServeMux) {
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Printf("Error ListenAndServe: %q", err.Error())
	}
}

func runDailyMessangers(cfg *config.Config, slackClient *slack.Client,
	dutyProvider processors.IDutyProvider, jiraClient processors.IJiraClient) {
	if cfg.DutyCommand.Enable {
		cron.AddJob(cron.EveryWorkingDayAt(cfg.DutyCommand.DailyMessageTime), &duty.DutyDailyMessenger{
			Config:       cfg,
			SlackClient:  slackClient,
			DutyProvider: dutyProvider,
		})
	}

	if cfg.TimelogsCommand.Enable {
		cron.AddJob(cron.EveryWorkingDayAt(cfg.TimelogsCommand.DailyMessageTime), &timelogs.TimelogsDailyMessenger{
			Config:      cfg,
			SlackClient: slackClient,
			JiraClient:  jiraClient,
		})
	}

	go cron.Run()
}

func main() {
	var configFilename string
	flag.StringVar(&configFilename, "config", "conf.yaml", "config file (yaml)")
	flag.Parse()

	cfg, err := config.ParseConfig(configFilename)
	if err != nil {
		log.Printf("Error parse config file %q: %s", configFilename, err.Error())
		return
	}

	slackClient := slack.NewClient(cfg.Slack.Token)
	cacheManager := cache.NewCache(DefaultCacheSize)
	jiraClient := jira.NewClient(cfg.Jira.Token)

	dutyProvider := opsgenie.NewOpsgenieClient(cfg.Opsgenie.Token)

	runDailyMessangers(cfg, slackClient, dutyProvider, jiraClient)

	mux := http.NewServeMux()
	commandProcessManager := initCommandProcessManager(cfg, slackClient, cacheManager, dutyProvider, jiraClient)
	initHandlers(mux, commandProcessManager)
	run(net.JoinHostPort(cfg.Main.Host, cfg.Main.Port), mux)
}
