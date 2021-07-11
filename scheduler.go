package main

import "time"

type PoolCommand struct {
	eCommand []CommandEntity
}

func (p *PoolCommand) Each(f func(entity *CommandEntity) bool, timeNow time.Time) {
	var buffer []CommandEntity
	for _, entity := range p.eCommand {
		res := false
		if entity.SendNow(timeNow) {
			res = f(&entity)
		}
		if res {
			entity.Sent(timeNow)
		}
		if !entity.IsOnce() || !res {
			buffer = append(buffer, entity)
		}
	}

	p.eCommand = buffer
}

func (p *PoolCommand) AddCommandEntity(entity CommandEntity) {
	p.eCommand = append(p.eCommand, entity)
}

type TypeSchedule string

const (
	Repeat   TypeSchedule = "timer"
	Schedule TypeSchedule = "schedule"
)

type CommandEntity interface {
	Command() ExecCommand
	SendNow(timeNow time.Time) bool
	Type() TypeSchedule
	IsOnce() bool
	Sent(timeNow time.Time)
}

type RepeatCommand struct {
	command  ExecCommand
	lastSend time.Time
	once     bool
	timer    time.Duration
}

func (r RepeatCommand) Command() ExecCommand {
	return r.command
}

func (r RepeatCommand) SendNow(timeNow time.Time) bool {
	return r.lastSend.Add(r.timer).Unix() <= timeNow.Unix()
}

func (r RepeatCommand) Type() TypeSchedule {
	return Repeat
}

func (r RepeatCommand) IsOnce() bool {
	return r.once
}

func (r *RepeatCommand) Sent(timeNow time.Time) {
	r.lastSend = timeNow
}

type ScheduleCommand struct {
	command  ExecCommand
	lastSend time.Time
	once     bool
	hmc      time.Duration
}

func (s ScheduleCommand) Command() ExecCommand {
	return s.command
}

func (s ScheduleCommand) Type() TypeSchedule {
	return Schedule
}

func (s ScheduleCommand) IsOnce() bool {
	return s.once
}

func (s *ScheduleCommand) Sent(timeNow time.Time) {
	s.lastSend = timeNow
}

func (s ScheduleCommand) SendNow(timeNow time.Time) bool {
	location := timeNow.Location()
	cYear, cMonth, cDay := timeNow.Date()
	lYear, lMonth, lDay := s.lastSend.Date()

	sNowDate := time.Date(cYear, cMonth, cDay, 0, 0, 0, 0, location)
	isBigStartDate := timeNow.Sub(sNowDate) < s.hmc
	lastNotEqualNow := lYear != cYear || lMonth != cMonth || lDay != cDay

	return isBigStartDate && lastNotEqualNow
}

type Command string

type ExecCommand interface {
	GetCommand() Command
	GetArgs() []string
}
