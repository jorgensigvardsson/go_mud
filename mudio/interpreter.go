package mudio

import (
	"strings"
)

type commandConstructor struct {
	name string
	cons func() Command
}

type CommandError struct {
	message string
}

func (e *CommandError) Error() string {
	return e.message
}

var unknownCommand = CommandError{"Unknown command."}

var commandConstructors = []commandConstructor{
	// Important: Keep the constructors sorted on name!
	// Also important: since we are matching user typed input as _prefix_ against
	// the command names, make sure to put shorter names before longer. For example,
	// put "north" before "nod" in the list, because the user is more likely to use directions
	// such as "n" (north) than using the nod emote.
	{name: "who", cons: NewCommandWho},
	{name: "quit", cons: NewCommandQuit},
}

type CommandParser = func(text string) (command Command, err error)

func ParseCommand(text string) (command Command, err error) {
	for _, commandConstructor := range commandConstructors {
		if strings.HasPrefix(commandConstructor.name, text) {
			return commandConstructor.cons(), nil
		}
	}

	return nil, &unknownCommand
}
