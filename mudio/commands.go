package mudio

import (
	"strings"

	"github.com/jorgensigvardsson/gomud/absmachine"
)

const InvalidInput = "Invalid input."

type CommandResponse struct {
	Player *absmachine.Player
	Text   string
}

type CommandResult struct {
	Prompt                 string
	TerminatationRequested bool
	Responses              []CommandResponse
	Output                 string
	TurnOffEcho            bool
	TurnOnEcho             bool
}

type Command interface {
	Execute(context *CommandContext) (result CommandResult, err *CommandError)
}

type CommandContext struct {
	Input  string
	World  *absmachine.World
	Player *absmachine.Player
}

type CommandError struct {
	message string
}

func (e *CommandError) Error() string {
	return e.message
}

func NewCommandError(message string) *CommandError {
	return &CommandError{message: message}
}

/**** Command: Who ****/
type CommandWho struct{}

func NewCommandWho(args []string) Command {
	return &CommandWho{}
}

func (command *CommandWho) Execute(context *CommandContext) (CommandResult, *CommandError) {
	b := buffer{}

	b.Println("Players On-line")
	b.Println("-------------------------------")

	for _, player := range context.Player.World.Players {
		suffix := ""
		if player == context.Player {
			suffix = " (You!)"
		}
		b.Printlnf("[%v] %v%v", player.Level, player.Name, suffix)
	}

	b.Println("-------------------------------")

	return CommandResult{Output: b.ToString()}, nil
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

func (command *CommandQuit) Execute(context *CommandContext) (CommandResult, *CommandError) {
	if len(command.args) > 0 && strings.ToLower(command.args[0]) == "now" {
		return CommandResult{Output: "Wow, what a hurry! Ok, sorry to see you go!\r\n", TerminatationRequested: true}, nil
	}

	if !command.isHandlingPrompt {
		// First execution, do nothing, but prompt user!
		command.isHandlingPrompt = true
		return CommandResult{Prompt: CommandQuitConfirmationMessage}, nil
	}

	// If we get here, we are handling the input from the prompt
	lcInput := strings.ToLower(context.Input)

	switch {
	case strings.HasPrefix("yes", lcInput):
		return CommandResult{Output: "Ok, sorry to see you go!\r\n", TerminatationRequested: true}, nil
	case strings.HasPrefix("no", lcInput):
		return CommandResult{}, nil
	default:
		return CommandResult{Prompt: InvalidInput}, nil
	}
}
