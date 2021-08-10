package mudio

// TODO: Break this file up into categories
// TODO: E.g.:
// TODO:   commands.go <- contains interface definitions
// TODO:   commands_movement.go <- contains movement commands
// TODO:   commands_login.go <- login/logout related commands
// TODO:   etc.

import (
	"strings"

	"github.com/jorgensigvardsson/gomud/absmachine"
)

type CommandError struct {
	message string
}

func (e *CommandError) Error() string {
	return e.message
}

var UnknownCommand = CommandError{"Unknown command."}
var InvalidInput = CommandError{"Invalid input."}

type Command interface {
	Execute(context *CommandContext) (CommandSubPrompter, error)
}

type CommandSubPrompter interface {
	Prompt(context *CommandContext) (string, error)
	ExecuteSubprompt(input string, context *CommandContext) (CommandSubPrompter, error)
}

type CommandQueue interface {
	Enqueue(command Command)
	Dequeue() Command
}

type CommandContext struct {
	World                *absmachine.World
	Player               *absmachine.Player
	Connection           TelnetConnection
	TerminationRequested bool
}

/**** Command: Who ****/
type CommandWho struct{}

func NewCommandWho() Command {
	return &CommandWho{}
}

func (command *CommandWho) Execute(context *CommandContext) (CommandSubPrompter, error) {
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
type CommandQuitSubPrompt struct{}

const CommandQuitConfirmationMessage = "Are you sure (y/n)?: "

func NewCommandQuit() Command {
	return &CommandQuit{}
}

func (command *CommandQuit) Execute(context *CommandContext) (CommandSubPrompter, error) {
	return command, nil
}

func (command *CommandQuit) Prompt(context *CommandContext) (string, error) {
	return CommandQuitConfirmationMessage, nil
}

func (command *CommandQuit) ExecuteSubprompt(input string, context *CommandContext) (CommandSubPrompter, error) {
	lcInput := strings.ToLower(input)

	switch {
	case strings.HasPrefix("yes", lcInput):
		context.Connection.WriteLine("Ok, sorry to see you go!")
		context.TerminationRequested = true
		return nil, nil
	case strings.HasPrefix("no", lcInput):
		return nil, nil
	default:
		return command, &InvalidInput
	}
}
