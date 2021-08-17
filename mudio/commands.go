package mudio

import (
	"strings"

	"github.com/jorgensigvardsson/gomud/absmachine"
	"github.com/jorgensigvardsson/gomud/logging"
)

const InvalidInput = "Invalid input."

type TextMessage struct {
	RecipientPlayer *absmachine.Player
	Text            string
}

type CommandResult struct {
	Prompt                 string
	TerminatationRequested bool
	TextMessages           []TextMessage
	Output                 string
	TurnOffEcho            bool
	TurnOnEcho             bool
}

type Command interface {
	Execute(context *CommandContext) (result CommandResult, err *CommandError)
}

type CommandRequirementsEvaluator func(player *absmachine.Player) bool

type CommandContext struct {
	Input  string
	World  *absmachine.World
	Player *absmachine.Player
	Logger logging.Logger
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

func NewCommandWho(args []string) (Command, CommandRequirementsEvaluator) {
	return &CommandWho{}, nil
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

func NewCommandQuit(args []string) (Command, CommandRequirementsEvaluator) {
	return &CommandQuit{args: args}, nil
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

func RequirePlayerLoggedIn(player *absmachine.Player) bool {
	return player.State.HasFlag(absmachine.PS_LOGGED_IN)
}

func RequirePlayerStanding(player *absmachine.Player) bool {
	return player.State.HasFlag(absmachine.PS_STANDING)
}

func CombineRequirements(evaluators ...CommandRequirementsEvaluator) CommandRequirementsEvaluator {
	return func(player *absmachine.Player) bool {
		for _, e := range evaluators {
			if !e(player) {
				return false
			}
		}
		return true
	}
}
