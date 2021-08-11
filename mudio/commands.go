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

const InvalidInput = "Invalid input."

var CommandFinished = CommandResult{}

func ContinueWithPrompt(prompt string) CommandResult {
	return CommandResult{
		Prompt:   prompt,
		Continue: true,
	}
}

type CommandResult struct {
	Prompt                 string
	Continue               bool
	TerminatationRequested bool
}

type Command interface {
	Execute(context *CommandContext) (result CommandResult, err error)
}

type CommandContext struct {
	Input      string
	World      *absmachine.World
	Player     *absmachine.Player
	Connection TelnetConnection
}

type CommandError struct {
	message string
}

func (e *CommandError) Error() string {
	return e.message
}

/**** Command: Who ****/
type CommandWho struct{}

func NewCommandWho(args []string) Command {
	return &CommandWho{}
}

func (command *CommandWho) Execute(context *CommandContext) (CommandResult, error) {
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

	return CommandFinished, nil
}

/**** Command: Quit ****/
type CommandQuit struct {
	args             []string
	isHandlingPrompt bool
}

const CommandQuitConfirmationMessage = "Are you sure (y/n)?: "

func NewCommandQuit(args []string) Command {
	return &CommandQuit{args: args}
}

func (command *CommandQuit) Execute(context *CommandContext) (CommandResult, error) {
	if len(command.args) > 0 && strings.ToLower(command.args[0]) == "now" {
		context.Connection.WriteLine("Wow, what a hurry! Ok, sorry to see you go!")
		return CommandResult{TerminatationRequested: true}, nil
	}

	if !command.isHandlingPrompt {
		// First execution, do nothing, but prompt user!
		command.isHandlingPrompt = true
		return ContinueWithPrompt(CommandQuitConfirmationMessage), nil
	}

	// If we get here, we are handling the input from the prompt
	lcInput := strings.ToLower(context.Input)

	switch {
	case strings.HasPrefix("yes", lcInput):
		context.Connection.WriteLine("Ok, sorry to see you go!")
		return CommandResult{TerminatationRequested: true}, nil
	case strings.HasPrefix("no", lcInput):
		return CommandResult{}, nil
	default:
		return ContinueWithPrompt(InvalidInput), nil
	}
}
