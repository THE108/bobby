package cron

import (
	"time"
)

const (
	defaultScheduleInterval time.Duration = 1 * time.Second
)

type ICronJob interface {
	Run()
}

type IChecker interface {
	Check(now time.Time) bool
}

type cronItem struct {
	job          ICronJob
	checker      IChecker
	lastExecTime time.Time
}

func (this cronItem) checkLastExecTime(now time.Time) bool {
	return this.lastExecTime.Add(24 * time.Hour).Before(now)
}

type Cron struct {
	scheduleInterval time.Duration
	jobs             []cronItem
	stop             chan struct{}
}

func NewCron(scheduleInterval time.Duration) *Cron {
	return &Cron{
		scheduleInterval: scheduleInterval,
		stop:             make(chan struct{}),
	}
}

func (this *Cron) AddJob(checker IChecker, job ICronJob) {
	this.jobs = append(this.jobs, cronItem{
		job:     job,
		checker: checker,
	})
}

func (this *Cron) Run() {
	if len(this.jobs) == 0 {
		return
	}

	for {
		select {
		case <-time.After(this.scheduleInterval):
			this.run()
		case <-this.stop:
			return
		}
	}
}

func (this *Cron) Stop() {
	this.stop <- struct{}{}
}

func (this *Cron) run() {
	now := time.Now()
	for i, item := range this.jobs {
		if item.checker.Check(now) && item.checkLastExecTime(now) {
			go item.job.Run()
			this.jobs[i].lastExecTime = now
		}
	}
}

var defaultCron *Cron = NewCron(defaultScheduleInterval)

func AddJob(checker IChecker, job ICronJob) {
	defaultCron.AddJob(checker, job)
}

func Run() {
	defaultCron.Run()
}

func Stop() {
	defaultCron.Stop()
}
