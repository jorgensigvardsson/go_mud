package mudio

import (
	"fmt"

	"github.com/jorgensigvardsson/gomud/absmachine"
)

/**** Command: Tell ****/
type CommandTell struct {
	args []string
}

func NewCommandTell(args []string) Command {
	return &CommandTell{args}
}

func (command *CommandTell) Execute(context *CommandContext) (CommandResult, *CommandError) {
	if len(command.args) == 0 {
		return CommandResult{}, &CommandError{"Who are you talking to?"}
	}

	otherPlayer := context.World.FindPlayerByName(command.args[0])

	if otherPlayer == context.Player {
		return CommandResult{}, &CommandError{"Talking to yourself??"}
	}

	if otherPlayer == nil {
		return CommandResult{}, &CommandError{fmt.Sprintf("Nobody with the name %v is online right now...", command.args[0])}
	}

	if otherPlayer.State.HasFlag(absmachine.PS_BUSY) {
		return CommandResult{}, &CommandError{fmt.Sprintf("%v is busy.", otherPlayer.Name)}
	}

	if len(command.args) == 1 {
		return CommandResult{}, &CommandError{fmt.Sprintf("Tell %v what?", otherPlayer.Name)}
	}

	// Ok, so we've done some basic checks here, and now we're ready to execute.
	// The biggest "hurdle" is that we've been handed an already parsed command line (command.args).
	// We want to send the text to otherPlayer that this player typed in verbatim. We have access to
	// the text in context.Input, so we'll extract the verbatim text there ourselves in this command!

	args, err := ParseArguments(context.Input, 2) // We want to split off the command and name (first two arguments of the command line)
	// The third argument is what the player wants to say

	if err != nil {
		return CommandResult{}, &CommandError{"Something went wrong here..."}
	}

	// We got the verbatim text, so let's push it to the other user!
	return CommandResult{
			Responses: []CommandResponse{
				{
					Player: otherPlayer,
					Text:   fmt.Sprintf("%v tells you: %v", context.Player.Name, args[2]),
				},
			},
		},
		nil
}
