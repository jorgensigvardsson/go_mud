package mudio

import (
	"sort"
)

/**** Command: Help ****/
type CommandHelp struct {
	args []string
}

func NewCommandHelp(args []string) (Command, CommandRequirementsEvaluator) {
	return &CommandHelp{args}, nil
}

func (command *CommandHelp) Execute(context *CommandContext) (CommandResult, *CommandError) {
	b := buffer{}

	if len(command.args) == 0 {
		// Give an index of all commands
		copy := commandConstructors

		// First sort on name...
		sort.Slice(copy, func(i, j int) bool {
			return copy[i].name < copy[j].name
		})

		// ...then on category
		sort.SliceStable(copy, func(i, j int) bool {
			return copy[i].cat < copy[j].cat
		})

		// Now we have commands sorted by name, and _grouped_ on category
		// because the sort was stable

		lastCat := ""
		for _, e := range copy {
			if lastCat != e.cat {
				b.Printlnf("$fg_yellow$..:: %v ::..$fg_white$", e.cat)
				lastCat = e.cat
			}

			b.Printf("%-15s", e.name)

			if e.shortDesc != "" {
				b.Printf(" %s", e.shortDesc)
			}

			b.Println("")
		}

	} else if len(command.args) == 1 {
		cmdCons := findCommandConstructor(command.args[0])

		if cmdCons == nil {
			return CommandResult{}, &CommandError{"There is no such command."}
		}

		if cmdCons.longDesc == "" {
			return CommandResult{}, &CommandError{"The command has no long description."}
		}

		b.Println(cmdCons.longDesc)
	} else {
		return CommandResult{}, &CommandError{"Huh?"}
	}

	return CommandResult{Output: b.ToString()}, nil
}
