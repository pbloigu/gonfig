package scheduler

import (
	"sync"

	"github.com/go-co-op/gocron/v2"
	"github.com/kelindar/event"
	"github.com/pbloigu/gonfig/api"
	"github.com/pbloigu/gonfig/server/events"
	"github.com/rs/zerolog/log"
)

type Scheduler interface {
	Start()
	RegisterCronTrigger(trigger api.CronTrigger)
	DeregisterCronTrigger(id int)
}

type sc struct {
	s           gocron.Scheduler
	triggerJobs map[int]gocron.Job
	jobLock     sync.RWMutex
}

func New() Scheduler {
	sc := sc{
		jobLock: sync.RWMutex{},
		s:       initScheduler(),
	}
	return &sc
}

func initScheduler() gocron.Scheduler {
	sc, err := gocron.NewScheduler()
	if err != nil {
		log.Panic().AnErr("error", err).Msg("Unable to create task scheduler.")
	}
	return sc
}

func (sc *sc) Start() {
	sc.s.Start()
}

func (sc *sc) DeregisterCronTrigger(id int) {
	sc.jobLock.Lock()
	defer sc.jobLock.Unlock()
	if j, ok := sc.triggerJobs[id]; ok {
		sc.s.RemoveJob(j.ID())
		delete(sc.triggerJobs, id)
	}
}

func (sc *sc) RegisterCronTrigger(trigger api.CronTrigger) {
	sc.jobLock.Lock()
	defer sc.jobLock.Unlock()
	jd := gocron.CronJob(trigger.CronExpression, false)
	task := gocron.NewTask(func() {
		event.Emit(events.Cron{
			Actions: trigger.Actions,
		})
	})
	j, err := sc.s.NewJob(jd, task)
	if err != nil {
		log.Panic().AnErr("error", err).Msg("Failed to register cron job.")
	}
	sc.triggerJobs[trigger.Id] = j
}
