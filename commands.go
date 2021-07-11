package main

import (
	"errors"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type ECommand struct {
	Com          Command
	FuncBuild    func(args []string) ExecCommand
	FuncExec     func(args []string, message tgbotapi.Message, pool *PoolCommand, tempo *Tempo) (string, bool, bool)
	FuncSchedule func(args []string, pool *PoolCommand, tempo *Tempo) (string, bool)
}

type PoolCommandHandlers struct {
	ECommand []ECommand
}

func (p PoolCommandHandlers) GetBuildHandlers() map[Command]ECommand {
	var buffer = make(map[Command]ECommand)
	for _, com := range p.ECommand {
		if com.FuncBuild != nil {
			buffer[com.Com] = com
		}
	}

	return buffer
}

func (p PoolCommandHandlers) GetExecHandlers() map[Command]ECommand {
	var buffer = make(map[Command]ECommand)
	for _, com := range p.ECommand {
		if com.FuncExec != nil {
			buffer[com.Com] = com
		}
	}

	return buffer
}

func (p PoolCommandHandlers) GetScheduleHandlers() map[Command]ECommand {
	var buffer = make(map[Command]ECommand)
	for _, com := range p.ECommand {
		if com.FuncSchedule != nil {
			buffer[com.Com] = com
		}
	}

	return buffer
}

func (p *PoolCommandHandlers) AddHandler(command ECommand) {
	p.ECommand = append(p.ECommand, command)
}

type BuildECommand struct {
	com  Command
	args []string
}

func (b BuildECommand) GetCommand() Command {
	return b.com
}

func (b BuildECommand) GetArgs() []string {
	return b.args
}

// BuildCommand build command from string
// @see CreateCommand
// @see ExecuteBuilder
func BuildCommand(com string, botName string) (Command, []string, error) {
	com = strings.ReplaceAll(com, "  ", " ")
	comArr := strings.Split(com, " ")
	if len(comArr) == 0 {
		return "", nil, errors.New("empty command")
	}

	commandString := strings.Replace(comArr[0], botName, "", 1)
	var args []string
	if len(comArr) > 1 {
		args = comArr[1:]
	}

	return Command(commandString), args, nil
}

// CreateCommand command from command
func CreateCommand(com Command, args []string, handlers *PoolCommandHandlers) ExecCommand {
	for c, fc := range handlers.GetBuildHandlers() {
		if c == com {
			return fc.FuncBuild(args)
		}
	}

	return nil
}

// ExecuteBuilder reply from command
func ExecuteBuilder(com Command, args []string, handlers PoolCommandHandlers, pool *PoolCommand, message tgbotapi.Message, tempo *Tempo) (string, bool, bool) {
	for c, fc := range handlers.GetExecHandlers() {
		if c == com {
			return fc.FuncExec(args, message, pool, tempo)
		}
	}

	return "", false, false
}

// ExecuteCommand message from command
func ExecuteCommand(command ExecCommand, handlers PoolCommandHandlers, pool *PoolCommand, tempo *Tempo) (string, bool) {
	for c, f := range handlers.GetScheduleHandlers() {
		if c == command.GetCommand() {
			return f.FuncSchedule(command.GetArgs(), pool, tempo)
		}
	}

	return "", false
}
