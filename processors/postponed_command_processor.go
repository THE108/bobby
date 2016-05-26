package processors

import (
	"log"
	"time"
)

type ResultProcessor interface {
	Init(args []string, now time.Time) error
	GetCacheKey() string
	Process() (string, error)
}

type ISlackPostponedClient interface {
	SendPostponedMessage(string, string) error
}

type ICache interface {
	Get(string) (string, bool)
	Set(string, string, time.Duration)
}

type PostponedCommandProcessor struct {
	SlackClient   ISlackPostponedClient
	Cache         ICache
	CacheDuration time.Duration
	Processor     ResultProcessor
	Token         string
}

func (this *PostponedCommandProcessor) GetAuthToken() string {
	return this.Token
}

func (this *PostponedCommandProcessor) ProcessCommand(command *SlackCommand, now time.Time, args []string) CommandResult {
	var text string
	if err := this.Processor.Init(args, now); err != nil {
		text += err.Error()
	}

	cacheKey := this.Processor.GetCacheKey()
	log.Printf("cache key: %s\n", cacheKey)
	if cachedText, found := this.Cache.Get(cacheKey); found {
		log.Printf("cachedText: %q, found: %v\n", cachedText, found)
		return CommandResult{
			Text: cachedText,
		}
	}

	go this.process(command, cacheKey, text)

	return CommandResult{
		Postponed: true,
	}
}

func (this *PostponedCommandProcessor) process(command *SlackCommand, cacheKey, text string) {
	processedText, err := this.Processor.Process()
	this.Cache.Set(cacheKey, processedText, this.CacheDuration)
	if err != nil {
		text += err.Error()
	}

	text += processedText

	if err := this.SlackClient.SendPostponedMessage(command.ResponseURL, text); err != nil {
		log.Printf("%s\n", err)
	}
}
