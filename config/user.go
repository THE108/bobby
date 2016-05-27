package config

type User struct {
	Name       string `yaml:"name"`
	JiraLogin  string `yaml:"jira-login"`
	SlackLogin string `yaml:"slack-login"`
}
