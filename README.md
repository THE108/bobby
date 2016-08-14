# Bobby Slack Bot

Slack bot helper for my team.

Features:
  * Duty notifications:
    * Who is on duty now?
    * Who is on duty next?
  * Timelogs notifications:
    * Who didn't log their work time
    
`Bobby Slack Bot` uses Jira and Opsgenie for data retrieval and Slack for personal and group notifications.

## Build

    cd bobby
    go install
    go build bobby
    
## Run

    ./bobby -config=conf.yaml

## Configuration

See `example_config.yaml`:

    main:
      host: 0.0.0.0
      port: 8080
    slack:
      token: <slack token>
      channel: <slack channel>
    jira:
      token: <jira token>
    opsgenie:
      token: <opsgenie token>
    pagerduty:
      token: <pagerduty token>
      subdomain: <your subdomain on pagerduty>
      timezone: Moscow
    duty-command:
      name: duty
      token: <slack auth token for duty command>
      schedule-ids:
        - <your schedule id>
      cache-ttl: 5m
      daily-message-time: 09:47
    timelogs-command:
      name: timelogs
      token: <slack auth token for timelogs command>
      minimum-time-logged: 6h
      cache-ttl: 5m
      daily-message-time: 09:47
      team:
      - name: "John Doe"
        jira-login: johndoe
        slack-login: john.doe
