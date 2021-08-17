package mudio

import (
	"strings"

	"github.com/jorgensigvardsson/gomud/absmachine"
)

type commandConstructor struct {
	name      string
	cat       string
	shortDesc string
	longDesc  string
	cons      func(args []string) (Command, CommandRequirementsEvaluator)
}

type CommandLine struct {
	Name string
	Args []string
}

var ErrUnknownCommand = &CommandError{"Unknown command."}
var ErrUnavailableCommand = &CommandError{"You can't do that right now."}
var ErrInvalidCommand = &CommandError{"Invalid command."}

const (
	CAT_Movement      = "Movement"
	CAT_Information   = "Information"
	CAT_Session       = "Session"
	CAT_Communication = "Communication"
)

var commandConstructors = []commandConstructor{
	// Important: Keep the constructors sorted on name!
	// Also important: since we are matching user typed input as _prefix_ against
	// the command names, make sure to put shorter names before longer. For example,
	// put "north" before "nod" in the list, because the user is more likely to use directions
	// such as "n" (north) than using the nod emote.

	// Directions - this are "prioritized"
	{name: "up", cons: NewCommandMoveUp, cat: CAT_Movement, shortDesc: "Moves character up"},
	{name: "down", cons: NewCommandMoveDown, cat: CAT_Movement, shortDesc: "Moves character down"},
	{name: "east", cons: NewCommandMoveEast, cat: CAT_Movement, shortDesc: "Moves character east"},
	{name: "west", cons: NewCommandMoveWest, cat: CAT_Movement, shortDesc: "Moves character west"},
	{name: "north", cons: NewCommandMoveNorth, cat: CAT_Movement, shortDesc: "Moves character north"},
	{name: "south", cons: NewCommandMoveSouth, cat: CAT_Movement, shortDesc: "Moves character south"},

	// Less prioritized commands
	{name: "help", cons: NewCommandHelp, cat: CAT_Information, shortDesc: "The manual!"},
	{name: "look", cons: NewCommandLook, cat: CAT_Information, shortDesc: "Allows for occular examination"},
	{name: "who", cons: NewCommandWho, cat: CAT_Session, shortDesc: "Who's online?"},
	{name: "quit", cons: NewCommandQuit, cat: CAT_Session, shortDesc: "For when you have to go!"},
	{name: "tell", cons: NewCommandTell, cat: CAT_Communication, shortDesc: "Send private messages to others"},
}

type CommandParser = func(text string, player *absmachine.Player) (command Command, err error)

func findCommandConstructor(text string) *commandConstructor {
	cmdNameLowerCase := strings.ToLower(text)
	for i, commandConstructor := range commandConstructors {
		if strings.HasPrefix(commandConstructor.name, cmdNameLowerCase) {
			return &commandConstructors[i]
		}
	}

	return nil
}

func ParseCommand(text string, player *absmachine.Player) (command Command, err error) {
	commandLine, err := ParseCommandLine(text)

	if err != nil {
		return nil, err
	}

	cmdCons := findCommandConstructor(commandLine.Name)

	if cmdCons == nil {
		return nil, ErrUnknownCommand
	}

	cmd, reqs := cmdCons.cons(commandLine.Args)

	if reqs != nil && !reqs(player) { // Does the command have requirements?
		return nil, ErrUnavailableCommand
	}

	return cmd, nil
}

func ParseCommandLine(text string) (CommandLine, error) {
	cmdEnd := strings.IndexAny(text, " \t")

	if cmdEnd < 0 {
		return CommandLine{Name: text}, nil
	}

	command := text[:cmdEnd]

	args, err := ParseArguments(strings.TrimSpace(text[cmdEnd+1:]), -1)

	if err != nil {
		return CommandLine{}, err
	}

	return CommandLine{Name: command, Args: args}, nil
}

// If count < 0, then the returned array will contain all arguments, neatly parsed.
// if count >= 0, then the returned array will contain count + 1 strings. The `count` first
// will be neatly parsed, and the last entry will contain the rest of the command line
func ParseArguments(text string, count int /* < 0 means parse ALL arguments */) ([]string, error) {
	start := 0
	insideQuotes := false

	args := make([]string, 0, strings.Count(text, " ") /* Estimate capacity */)

	var i int
	for i = 0; i < len(text) && (count < 0 || len(args) < count); i++ {
		if text[i] == ' ' || text[i] == '\t' {
			if insideQuotes {
				// Let whitespace become part of the argument (inside quotes!)
			} else {
				// Now we have an argument!
				if i > start {
					args = append(args, text[start:i])
				}
				start = i + 1
			}
		} else if text[i] == '"' {
			if insideQuotes {
				if text[i-1] == '\\' {
					// Escaped " - let become part of the argument
				} else {
					// Now we have an argument!
					args = append(args, strings.ReplaceAll(text[start+1:i], "\\\"", "\"")) // Make sure escaped quotes are turned into just quotes!
					start = i + 1
					insideQuotes = false
				}
			} else {
				insideQuotes = true
				start = i
			}
		}
	}

	if count < 0 { // We want all arguments
		if insideQuotes {
			return args, ErrInvalidCommand
		} else {
			args = append(args, text[start:i])
		}
	} else {
		// Trim off all whitspace if any
		for ; start < len(text) && (text[start:][0] == ' ' || text[start:][0] == '\t'); start++ {
		}
		args = append(args, text[start:])
	}

	return args, nil
}
