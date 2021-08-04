package mudio

import (
	"github.com/jorgensigvardsson/gomud/absmachine"
)

type CommandError struct {
	message string
}

func (e *CommandError) Error() string {
	return e.message
}

var UnknownCommand = &CommandError{"Unknown command"}

type Command interface {
	Execute(context *CommandContext) (CommandSubPrompter, *CommandError)
}

type CommandSubPrompter interface {
	Prompt() string
	GiveInput(input string, context *CommandContext) (CommandSubPrompter, *CommandError)
}

type CommandQueue interface {
	Enqueue(command Command)
	Dequeue() Command
}

type CommandContext struct {
	Player     *absmachine.Player
	Connection TelnetConnection
}

/**** Command: Who ****/
type CommandWho struct{}

func NewCommandWho() Command {
	return &CommandWho{}
}

func (command *CommandWho) Execute(context *CommandContext) (CommandSubPrompter, *CommandError) {
	conn := context.Connection

	conn.WriteLine("Players On-line")
	conn.WriteLine("-------------------------------")

	for _, player := range context.Player.World.Players {
		suffix := ""
		if player == context.Player {
			suffix = " (You!)"
		}
		conn.WriteLinef("[%v] %v%v", player.Level, player.Name, suffix)
	}

	conn.WriteLine("-------------------------------")

	return nil, nil
}

/**** Command: Quit ****/
type CommandQuit struct{}

func NewCommandQuit() Command {
	return &CommandQuit{}
}

func (command *CommandQuit) Execute(context *CommandContext) (CommandSubPrompter, *CommandError) {
	return nil, &CommandError{message: "Not implemented"}
}
