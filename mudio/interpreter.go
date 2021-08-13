package mudio

import (
	"errors"
	"strings"
)

type commandConstructor struct {
	name string
	cons func(args []string) Command
}

type CommandLine struct {
	Name string
	Args []string
}

var ErrUnknownCommand = errors.New("unknown command")

var commandConstructors = []commandConstructor{
	// Important: Keep the constructors sorted on name!
	// Also important: since we are matching user typed input as _prefix_ against
	// the command names, make sure to put shorter names before longer. For example,
	// put "north" before "nod" in the list, because the user is more likely to use directions
	// such as "n" (north) than using the nod emote.
	{name: "look", cons: NewCommandLook},
	{name: "who", cons: NewCommandWho},
	{name: "quit", cons: NewCommandQuit},
	{name: "tell", cons: NewCommandTell},
}

type CommandParser = func(text string) (command Command, err error)

func ParseCommand(text string) (command Command, err error) {
	commandLine, err := ParseCommandLine(text)

	if err != nil {
		return nil, err
	}

	cmdNameLowerCase := strings.ToLower(commandLine.Name)
	for _, commandConstructor := range commandConstructors {
		if strings.HasPrefix(commandConstructor.name, cmdNameLowerCase) {
			return commandConstructor.cons(commandLine.Args), nil
		}
	}

	return nil, ErrUnknownCommand
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

var ErrInvalidCommandLine = errors.New("invalid command line")

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
			return args, ErrInvalidCommandLine
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
