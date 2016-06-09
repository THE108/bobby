package processors

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

type CommandResult struct {
	Text      string
	Postponed bool
}

type ICommandProcessor interface {
	ProcessCommand(command *SlackCommand, now time.Time, args []string) CommandResult
	GetAuthToken() string
}

type CommandProcessManager struct {
	lock       sync.RWMutex
	processors map[string]ICommandProcessor
}

func NewCommandProcessManager() *CommandProcessManager {
	return &CommandProcessManager{
		processors: make(map[string]ICommandProcessor, 2),
	}
}

func (this *CommandProcessManager) AddCommandProcessor(commandName string, processor ICommandProcessor) {
	this.lock.Lock()
	this.processors[commandName] = processor
	this.lock.Unlock()
}

func (this *CommandProcessManager) ProcessCommand(command *SlackCommand) (result CommandResult, err error) {
	commandName := strings.Trim(command.Command, "/ ")
	if len(commandName) == 0 {
		return result, fmt.Errorf("empty command")
	}

	this.lock.RLock()
	commandProcessor, found := this.processors[commandName]
	this.lock.RUnlock()

	if !found {
		return result, fmt.Errorf("unknown command %q", commandName)
	}

	if command.Token != commandProcessor.GetAuthToken() {
		return result, fmt.Errorf("validation failed: invalid token %q", command.Token)
	}

	var args []string
	command.Text = strings.Trim(command.Text, "/ ")
	if len(command.Text) != 0 {
		args = strings.Split(command.Text, " ")
	}

	log.Printf("args: %+v len(%d)\n", args, len(args))

	return commandProcessor.ProcessCommand(command, time.Now(), args), nil
}
