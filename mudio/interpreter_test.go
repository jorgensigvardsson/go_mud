package mudio

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/jorgensigvardsson/gomud/absmachine"
)

func Test_EmptyInput_EmptyCommand(t *testing.T) {
	// Arrange
	const input = ""

	// Act
	result, err := ParseCommandLine(input)

	if err != nil {
		t.Errorf("Did not expect error %v", err)
	}

	// Assert
	if result.Name != "" {
		t.Errorf("Did not expect name %v", result.Name)
	}

	if len(result.Args) != 0 {
		t.Errorf("Did not expect arguments %v", result.Args)
	}
}

func Test_SimpleCommand_NoArgs(t *testing.T) {
	// Arrange
	const input = "command"

	// Act
	result, err := ParseCommandLine(input)

	if err != nil {
		t.Errorf("Did not expect error %v", err)
	}

	// Assert
	if result.Name != "command" {
		t.Errorf("Did not expect name %v", result.Name)
	}

	if len(result.Args) != 0 {
		t.Errorf("Did not expect arguments %v", result.Args)
	}
}

func Test_SimpleCommand_SimpleArgs(t *testing.T) {
	// Arrange
	const input = "command a b c"

	// Act
	result, err := ParseCommandLine(input)

	if err != nil {
		t.Errorf("Did not expect error %v", err)
	}

	// Assert
	if result.Name != "command" {
		t.Errorf("Did not expect name %v", result.Name)
	}

	if len(result.Args) != 3 {
		t.Errorf("Did not expect arguments %v", result.Args)
	} else {
		if result.Args[0] != "a" {
			t.Errorf("Did not expect argument 0 to be %v", result.Args[0])
		}

		if result.Args[1] != "b" {
			t.Errorf("Did not expect argument 1 to be %v", result.Args[1])
		}

		if result.Args[2] != "c" {
			t.Errorf("Did not expect argument 2 to be %v", result.Args[2])
		}
	}
}

func Test_SimpleCommand_ComplexArgs(t *testing.T) {
	// Arrange
	const input = "command a \"b c\" d"

	// Act
	result, err := ParseCommandLine(input)

	if err != nil {
		t.Errorf("Did not expect error %v", err)
	}

	// Assert
	if result.Name != "command" {
		t.Errorf("Did not expect name %v", result.Name)
	}

	if len(result.Args) != 3 {
		t.Errorf("Did not expect arguments %v", result.Args)
	} else {
		if result.Args[0] != "a" {
			t.Errorf("Did not expect argument 0 to be %v", result.Args[0])
		}

		if result.Args[1] != "b c" {
			t.Errorf("Did not expect argument 1 to be %v", result.Args[1])
		}

		if result.Args[2] != "d" {
			t.Errorf("Did not expect argument 2 to be %v", result.Args[2])
		}
	}
}

func Test_SimpleCommand_QuotedQuotes(t *testing.T) {
	// Arrange
	const input = "command a \"b\\\"c\" d"

	// Act
	result, err := ParseCommandLine(input)

	if err != nil {
		t.Errorf("Did not expect error %v", err)
	}

	// Assert
	if result.Name != "command" {
		t.Errorf("Did not expect name %v", result.Name)
	}

	if len(result.Args) != 3 {
		t.Errorf("Did not expect arguments %v", result.Args)
	} else {
		if result.Args[0] != "a" {
			t.Errorf("Did not expect argument 0 to be %v", result.Args[0])
		}

		if result.Args[1] != "b\"c" {
			t.Errorf("Did not expect argument 1 to be %v", result.Args[1])
		}

		if result.Args[2] != "d" {
			t.Errorf("Did not expect argument 2 to be %v", result.Args[2])
		}
	}
}

func Test_SimpleCommand_UnclosedQuote(t *testing.T) {
	// Arrange
	const input = "command a \"b c d"

	// Act
	_, err := ParseCommandLine(input)

	// Assert
	if err != ErrInvalidCommand {
		t.Errorf("Did not expect error %v", err)
	}
}

type commandTestTableEntry struct {
	commandInput    string
	commandTypeName string
}

var commandTestTable []commandTestTableEntry = []commandTestTableEntry{
	{
		commandInput:    "who",
		commandTypeName: "CommandWho",
	},
}

func Test_ParseCommand(t *testing.T) {
	for _, e := range commandTestTable {
		command, err := ParseCommand(e.commandInput, &absmachine.Player{})

		if err != nil {
			t.Errorf("Command input \"%v\" generated the error: %v", e.commandInput, err)
		} else {
			commandType := reflect.TypeOf(command)
			var commandName string
			if commandType.Kind() == reflect.Ptr {
				commandName = commandType.Elem().Name()
			} else {
				commandName = commandType.Name()
			}

			if commandName != e.commandTypeName {
				t.Errorf("Command input \"%v\" generated a command object of type %v", e.commandInput, commandName)
			}

		}
	}
}

func Test_ParseCommand_InvalidCommandLine(t *testing.T) {
	_, err := ParseCommand("invalid \"command line", &absmachine.Player{})
	if err != ErrInvalidCommand {
		t.Errorf("Command input generated the unexpected error: %v", err)
	}
}

func Test_ParseCommand_UnknownCommand(t *testing.T) {
	_, err := ParseCommand("garbledigarbage", &absmachine.Player{})
	if err != ErrUnknownCommand {
		t.Errorf("Command input generated the unexpected error: %v", err)
	}
}

type parseArgsTestCase struct {
	commandLine  string
	count        int
	expectedArgs []string
}

var parseArgsTestCases = []parseArgsTestCase{
	{commandLine: "this is a test", count: -1, expectedArgs: []string{"this", "is", "a", "test"}},
	{commandLine: "this is a test", count: 0, expectedArgs: []string{"this is a test"}},
	{commandLine: "this is a test", count: 1, expectedArgs: []string{"this", "is a test"}},
	{commandLine: "this is a test", count: 2, expectedArgs: []string{"this", "is", "a test"}},
	{commandLine: "this is a test", count: 3, expectedArgs: []string{"this", "is", "a", "test"}},
	{commandLine: "this is a test", count: 4, expectedArgs: []string{"this", "is", "a", "test"}},
	{commandLine: "this is a test", count: 100, expectedArgs: []string{"this", "is", "a", "test"}},
	{commandLine: "this \"is a\" test", count: -1, expectedArgs: []string{"this", "is a", "test"}},
	{commandLine: "this \"is a\" test", count: 0, expectedArgs: []string{"this \"is a\" test"}},
	{commandLine: "this \"is a\" test", count: 1, expectedArgs: []string{"this", "\"is a\" test"}},
	{commandLine: "this \"is a\" test", count: 2, expectedArgs: []string{"this", "is a", "test"}},
}

func Test_ParseArguments(t *testing.T) {
	for _, testCase := range parseArgsTestCases {
		t.Run(
			fmt.Sprintf("%v, count = %v", testCase.commandLine, testCase.count),
			func(t *testing.T) {
				args, err := ParseArguments(testCase.commandLine, testCase.count)

				if err != nil {
					t.Error("Did not expect an error:", err)
				}

				if len(args) != len(testCase.expectedArgs) {
					t.Error("Returned array is not what is expected", args)
				} else {
					for i, expectedArg := range testCase.expectedArgs {
						if expectedArg != args[i] {
							t.Error("Returned array is not what is expected", args)
						}
					}
				}
			},
		)
	}
}
