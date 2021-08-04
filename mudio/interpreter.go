package mudio

import (
	"sort"
	"strings"
)

type commandConstructor struct {
	name string
	cons func() Command
}

var commandConstructors = []commandConstructor{
	{name: "who", cons: NewCommandWho},
	{name: "quit", cons: NewCommandQuit},
}

func InitializeInterpreter() {
	sort.Slice(
		commandConstructors,
		func(i, j int) bool {
			return commandConstructors[i].name < commandConstructors[j].name
		},
	)
}

func Parse(text string) (command Command, err *CommandError) {
	for _, commandConstructor := range commandConstructors {
		if strings.HasPrefix(commandConstructor.name, text) {
			return commandConstructor.cons(), nil
		}
	}

	return nil, UnknownCommand
}
