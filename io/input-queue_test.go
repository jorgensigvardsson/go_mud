package io

import (
	"fmt"
	"testing"

	"github.com/jorgensigvardsson/gomud/absmachine"
	"github.com/jorgensigvardsson/gomud/logging"
	"github.com/jorgensigvardsson/gomud/mudio"
)

func Test_showNormalPrompt(t *testing.T) {
	player := absmachine.Player{
		Health: 103,
		Mana:   43,
	}

	promptText := normalPrompt(&player)

	if promptText != "[H:103] [M:43] > " {
		t.Errorf("Unexpected prompt: %v", promptText)
	}
}

func Test_Append_PlayerLimitIsRespected(t *testing.T) {
	q := NewInputQueue(1, 1, logging.NewNullLogger())
	p1 := absmachine.NewPlayer()
	p2 := absmachine.NewPlayer()
	outputChannel1 := make(chan *PlayerOutput, 10)
	outputChannel2 := make(chan *PlayerOutput, 10)
	errorChannel1 := make(chan error, 10)
	errorChannel2 := make(chan error, 10)

	p1.Name = "p1"
	p2.Name = "p2"

	q.Append(
		&PlayerInput{
			player:             p1,
			text:               "cmd",
			outputChannel:      outputChannel1,
			errorReturnChannel: errorChannel1,
		},
	)

	q.Append(
		&PlayerInput{
			player:             p2,
			text:               "cmd",
			outputChannel:      outputChannel2,
			errorReturnChannel: errorChannel2,
		},
	)

	select {
	case err := <-errorChannel1:
		t.Errorf("Unexpected error on channel 1: %v", err)
	default:
	}

	select {
	case err := <-errorChannel2:
		if err != ErrTooManyPlayers {
			t.Errorf("Unexpected error: %v", err)
		}
	default:
		t.Errorf("Unexpectedly, there was no error on channel 2!")
	}
}

func Test_Append_PlayerInputLimitIsRespected(t *testing.T) {
	q := NewInputQueue(1, 1, logging.NewNullLogger())
	p := absmachine.NewPlayer()
	outputChannel := make(chan *PlayerOutput, 10)
	errorChannel := make(chan error, 10)

	p.Name = "p1"

	q.Append(
		&PlayerInput{
			player:             p,
			text:               "cmd 1",
			errorReturnChannel: errorChannel,
			outputChannel:      outputChannel,
		},
	)

	q.Append(
		&PlayerInput{
			player:             p,
			text:               "cmd 2",
			errorReturnChannel: errorChannel,
			outputChannel:      outputChannel,
		},
	)

	select {
	case err := <-errorChannel:
		if err != ErrTooMuchInput {
			t.Errorf("Unexpected error: %v", err)
		}
	default:
		t.Errorf("Unexpectedly, there was no error on channel!")
	}
}

func Test_Execute_NoInput_NoEffect(t *testing.T) {
	// Arrange
	q := NewInputQueue(10, 10, logging.NewNullLogger())
	player := absmachine.NewPlayer()
	world := absmachine.NewWorld()

	world.AddPlayers([]*absmachine.Player{player})

	// Act
	q.Execute(world)

	// Assert
	if len(world.Players) != 1 {
		t.Errorf("Number of players in world has changed from expected count 1: %v", len(world.Players))
	}

	if world.Players[0] != player {
		t.Error("Expected player not in world:", world.Players[0])
	}
}

func Test_Execute_PlayersHaveBeenAdded_ButHasNoInput_NoEffect(t *testing.T) {
	// Arrange
	q := NewInputQueue(10, 10, logging.NewNullLogger())
	player := absmachine.NewPlayer()
	world := absmachine.NewWorld()

	world.AddPlayers([]*absmachine.Player{player})
	q.playerQueues[player] = newPlayerQueue()

	// Act
	q.Execute(world)

	// Assert
	if len(world.Players) != 1 {
		t.Errorf("Number of players in world has changed from expected count 1: %v", len(world.Players))
	}

	if world.Players[0] != player {
		t.Error("Expected player not in world:", world.Players[0])
	}
}

func Test_Execute_PlayerHasEvent_PE_Exited_PlayerIsRemovedFromWorld(t *testing.T) {
	// Arrange
	q := NewInputQueue(10, 10, logging.NewNullLogger())
	player := absmachine.NewPlayer()
	world := absmachine.NewWorld()

	world.AddPlayers([]*absmachine.Player{player})
	q.Append(&PlayerInput{
		player:             player,
		event:              PE_Exited,
		errorReturnChannel: make(chan<- error, 1),
		outputChannel:      make(chan<- *PlayerOutput),
	})

	// Act
	q.Execute(world)

	// Assert
	if len(world.Players) != 0 {
		t.Errorf("Number of players in world has changed from expected count 0: %v", len(world.Players))
	}
}

func Test_Execute_PlayerHasEvent_UnknownEvent_NoEffect(t *testing.T) {
	// Arrange
	q := NewInputQueue(10, 10, logging.NewNullLogger())
	player := absmachine.NewPlayer()
	world := absmachine.NewWorld()

	world.AddPlayers([]*absmachine.Player{player})
	q.Append(&PlayerInput{
		player:             player,
		event:              PE_EventCount + 1,
		errorReturnChannel: make(chan<- error, 1),
		outputChannel:      make(chan<- *PlayerOutput),
	})

	// Act
	q.Execute(world)

	// Assert
	if len(world.Players) != 1 {
		t.Errorf("Number of players in world has changed from expected count 1: %v", len(world.Players))
	}

	if world.Players[0] != player {
		t.Error("Expected player not in world:", world.Players[0])
	}
}

type FakeCommand struct {
	receivedContext *mudio.CommandContext
	returnError     *mudio.CommandError
	returnResult    mudio.CommandResult
}

func (cmd *FakeCommand) Execute(context *mudio.CommandContext) (result mudio.CommandResult, err *mudio.CommandError) {
	cmd.receivedContext = context
	return cmd.returnResult, cmd.returnError
}

func getOutput(channel <-chan *PlayerOutput) []*PlayerOutput {
	output := make([]*PlayerOutput, 0)
	done := false

	for !done {
		select {
		case playerOutput := <-channel:
			output = append(output, playerOutput)
		default:
			done = true
		}
	}

	return output
}

func Test_Execute_NoInput_StandardPromptWrittenToConnection(t *testing.T) {
	// Arrange
	q := NewInputQueue(10, 10, logging.NewNullLogger())
	player := absmachine.NewPlayer()
	world := absmachine.NewWorld()
	outputChannel := make(chan *PlayerOutput, 10)

	world.AddPlayers([]*absmachine.Player{player})

	q.Append(&PlayerInput{
		player:             player,
		errorReturnChannel: make(chan<- error, 1),
		outputChannel:      outputChannel,
	})

	// Act
	q.Execute(world)

	// Assert
	output := getOutput(outputChannel)

	if len(output) != 1 || output[0].text != "[H:0] [M:0] > " {
		t.Error("Unexpected output sent to player")
	}
}

func Test_Execute_PlayerHasCurrentCommand_InputIsSentToCurrentCommand_CommandFinished(t *testing.T) {
	// Arrange
	q := NewInputQueue(10, 10, logging.NewNullLogger())
	player := absmachine.NewPlayer()
	world := absmachine.NewWorld()
	outputChannel := make(chan *PlayerOutput, 10)
	currentCommand := FakeCommand{
		returnError:  nil,
		returnResult: mudio.CommandResult{},
	}

	world.AddPlayers([]*absmachine.Player{player})
	q.playerQueues[player] = newPlayerQueue()
	q.playerQueues[player].currentCommand = &currentCommand

	q.Append(&PlayerInput{
		player:             player,
		text:               "cmd text",
		errorReturnChannel: make(chan<- error, 1),
		outputChannel:      outputChannel,
	})

	// Act
	q.Execute(world)

	// Assert
	if currentCommand.receivedContext.Input != "cmd text" {
		t.Errorf("Command did not expect the input: %v", currentCommand.receivedContext.Input)
	}

	if q.playerQueues[player].currentCommand != nil {
		t.Error("Expected current command for player to be cleared!")
	}
}

func Test_Execute_PlayerHasCurrentCommand_InputIsSentToCurrentCommand_CommandErrorsAreWrittenToConnection(t *testing.T) {
	// Arrange
	q := NewInputQueue(10, 10, logging.NewNullLogger())
	player := absmachine.NewPlayer()
	world := absmachine.NewWorld()
	outputChannel := make(chan *PlayerOutput, 10)
	currentCommand := FakeCommand{
		returnError:  mudio.NewCommandError("foo"),
		returnResult: mudio.CommandResult{},
	}

	world.AddPlayers([]*absmachine.Player{player})
	q.playerQueues[player] = newPlayerQueue()
	q.playerQueues[player].currentCommand = &currentCommand

	q.Append(&PlayerInput{
		player:             player,
		text:               "cmd text",
		errorReturnChannel: make(chan<- error, 1),
		outputChannel:      outputChannel,
	})

	// Act
	q.Execute(world)

	// Assert
	if currentCommand.receivedContext.Input != "cmd text" {
		t.Errorf("Command did not expect the input: %v", currentCommand.receivedContext.Input)
	}

	if q.playerQueues[player].currentCommand != nil {
		t.Error("Expected current command for player to be cleared!")
	}

	output := getOutput(outputChannel)

	if len(output) != 2 {
		t.Errorf("Expected 2 output item, but got %v", len(output))
	} else {
		if output[0].text != fmt.Sprintln("foo") {
			t.Error("Unexpected output sent to player")
		}

		if output[1].text != "[H:0] [M:0] > " {
			t.Error("Unexpected output sent to player")
		}
	}
}

/*
func Test_Execute_PlayerHasCurrentCommand_InputIsSentToCurrentCommand_CommandWantsToContinue(t *testing.T) {
	// Arrange
	conn := &FakeTelnetConnection{}
	q := NewInputQueue(10, 10, logging.NewNullLogger())
	player := absmachine.NewPlayer()
	world := absmachine.NewWorld()
	currentCommand := FakeCommand{
		returnError:  nil,
		returnResult: mudio.ContinueWithPrompt("A prompt"),
	}

	world.AddPlayers([]*absmachine.Player{player})
	q.playerQueues[player] = newPlayerQueue()
	q.playerQueues[player].currentCommand = &currentCommand

	q.Append(&PlayerInput{
		player:             player,
		text:               "cmd text",
		connection:         conn,
		errorReturnChannel: make(chan<- error, 1),
	})

	// Act
	q.Execute(world)

	// Assert
	if currentCommand.receivedContext.Input != "cmd text" {
		t.Errorf("Command did not expect the input: %v", currentCommand.receivedContext.Input)
	}

	if q.playerQueues[player].currentCommand != &currentCommand {
		t.Error("Expected current command for player to remain!")
	}

	if conn.writtenText != "A prompt" {
		t.Errorf("Unexpected output sent to player Connection: %v", conn.writtenText)
	}
}

func Test_Execute_PlayerHasCurrentCommand_InputIsSentToCurrentCommand_CommandWantsToTerminate(t *testing.T) {
	// Arrange
	conn := &FakeTelnetConnection{}
	q := NewInputQueue(10, 10, logging.NewNullLogger())
	player := absmachine.NewPlayer()
	world := absmachine.NewWorld()
	currentCommand := FakeCommand{
		returnError: nil,
		returnResult: mudio.CommandResult{
			TerminatationRequested: true,
		},
	}
	errorReturnChannel := make(chan error, 10)

	world.AddPlayers([]*absmachine.Player{player})
	q.playerQueues[player] = newPlayerQueue()
	q.playerQueues[player].currentCommand = &currentCommand

	q.Append(&PlayerInput{
		player:             player,
		text:               "cmd text",
		connection:         conn,
		errorReturnChannel: errorReturnChannel,
	})

	// Act
	q.Execute(world)

	// Assert
	if currentCommand.receivedContext.Input != "cmd text" {
		t.Errorf("Command did not expect the input: %v", currentCommand.receivedContext.Input)
	}

	if q.playerQueues[player].currentCommand != nil {
		t.Error("Expected current command for player to be cleared!")
	}

	select {
	case err := <-errorReturnChannel:
		if err != ErrPlayerQuit {
			t.Errorf("Did not expect '%v' on the error return channel", err)
		}
	default:
		t.Error("Expected an error on the error return channel!")
	}
}

func Test_Execute_PlayerHasNoCurrentCommand_InvalidInput_ErrorAndStandardPromptWrittenToConnection(t *testing.T) {
	// Arrange
	conn := &FakeTelnetConnection{}
	q := NewInputQueue(10, 10, logging.NewNullLogger())
	player := absmachine.NewPlayer()
	world := absmachine.NewWorld()

	q.commandParser = func(text string) (command mudio.Command, err error) {
		return nil, errors.New("foo")
	}

	world.AddPlayers([]*absmachine.Player{player})

	q.Append(&PlayerInput{
		player:             player,
		text:               "gargksjdl",
		connection:         conn,
		errorReturnChannel: make(chan<- error, 1),
	})

	// Act
	q.Execute(world)

	// Assert
	if conn.writtenText != "Error: foo\r\n[H:0] [M:0] > " {
		t.Errorf("Unexpected output sent to player Connection: %v", conn.writtenText)
	}
}

func Test_Execute_PlayerHasNoCurrentCommand_InputIsSentToParsedCommand_CommandFinished(t *testing.T) {
	// Arrange
	conn := &FakeTelnetConnection{}
	q := NewInputQueue(10, 10, logging.NewNullLogger())
	player := absmachine.NewPlayer()
	world := absmachine.NewWorld()
	fakeCommand := FakeCommand{
		returnError:  nil,
		returnResult: mudio.CommandResult{},
	}

	q.commandParser = func(text string) (command mudio.Command, err error) {
		return &fakeCommand, nil
	}

	world.AddPlayers([]*absmachine.Player{player})

	q.Append(&PlayerInput{
		player:             player,
		text:               "cmd text",
		connection:         conn,
		errorReturnChannel: make(chan<- error, 1),
	})

	// Act
	q.Execute(world)

	// Assert
	if fakeCommand.receivedContext == nil {
		t.Error("Fake command was never called?")
	}

	if q.playerQueues[player].currentCommand != nil {
		t.Error("Expected current command for player to be cleared!")
	}

	if conn.writtenText != "[H:0] [M:0] > " {
		t.Errorf("Unexpected output sent to player Connection: %v", conn.writtenText)
	}
}

func Test_Execute_PlayerHasNoCurrentCommand_InputIsSentToParsedCommand_CommandErrorsAreWrittenToConnection(t *testing.T) {
	// Arrange
	conn := &FakeTelnetConnection{}
	q := NewInputQueue(10, 10, logging.NewNullLogger())
	player := absmachine.NewPlayer()
	world := absmachine.NewWorld()
	fakeCommand := FakeCommand{
		returnError:  mudio.NewCommandError("foo"),
		returnResult: mudio.CommandResult{},
	}

	q.commandParser = func(text string) (command mudio.Command, err error) {
		return &fakeCommand, nil
	}

	world.AddPlayers([]*absmachine.Player{player})

	q.Append(&PlayerInput{
		player:             player,
		text:               "cmd text",
		connection:         conn,
		errorReturnChannel: make(chan<- error, 1),
	})

	// Act
	q.Execute(world)

	// Assert
	if fakeCommand.receivedContext == nil {
		t.Error("Fake command was never called?")
	}

	if q.playerQueues[player].currentCommand != nil {
		t.Error("Expected current command for player to be cleared!")
	}

	if conn.writtenText != "foo\r\n[H:0] [M:0] > " {
		t.Errorf("Unexpected output sent to player Connection: %v", conn.writtenText)
	}
}

func Test_Execute_PlayerHasNoCurrentCommand_InputIsSentToParsedCommand_CommandWantsToContinue(t *testing.T) {
	// Arrange
	conn := &FakeTelnetConnection{}
	q := NewInputQueue(10, 10, logging.NewNullLogger())
	player := absmachine.NewPlayer()
	world := absmachine.NewWorld()
	fakeCommand := FakeCommand{
		returnError:  nil,
		returnResult: mudio.ContinueWithPrompt("A prompt"),
	}

	q.commandParser = func(text string) (command mudio.Command, err error) {
		return &fakeCommand, nil
	}

	world.AddPlayers([]*absmachine.Player{player})

	q.Append(&PlayerInput{
		player:             player,
		text:               "cmd text",
		connection:         conn,
		errorReturnChannel: make(chan<- error, 1),
	})

	// Act
	q.Execute(world)

	// Assert
	if fakeCommand.receivedContext == nil {
		t.Error("Fake command was never called?")
	}

	if q.playerQueues[player].currentCommand != &fakeCommand {
		t.Error("Expected current command for player to remain!")
	}

	if conn.writtenText != "A prompt" {
		t.Errorf("Unexpected output sent to player Connection: %v", conn.writtenText)
	}
}

func Test_Execute_PlayerHasNoCurrentCommand_InputIsSentToParsedCommand_CommandWantsToTerminate(t *testing.T) {
	// Arrange
	conn := &FakeTelnetConnection{}
	q := NewInputQueue(10, 10, logging.NewNullLogger())
	player := absmachine.NewPlayer()
	world := absmachine.NewWorld()
	fakeCommand := FakeCommand{
		returnError: nil,
		returnResult: mudio.CommandResult{
			TerminatationRequested: true,
		},
	}
	errorReturnChannel := make(chan error, 10)

	q.commandParser = func(text string) (command mudio.Command, err error) {
		return &fakeCommand, nil
	}

	world.AddPlayers([]*absmachine.Player{player})

	q.Append(&PlayerInput{
		player:             player,
		text:               "cmd text",
		connection:         conn,
		errorReturnChannel: errorReturnChannel,
	})

	// Act
	q.Execute(world)

	// Assert
	if fakeCommand.receivedContext == nil {
		t.Error("Fake command was never called?")
	}

	if q.playerQueues[player].currentCommand != nil {
		t.Error("Expected current command for player to be cleared!")
	}

	select {
	case err := <-errorReturnChannel:
		if err != ErrPlayerQuit {
			t.Errorf("Did not expect '%v' on the error return channel", err)
		}
	default:
		t.Error("Expected an error on the error return channel!")
	}
}

func Test_Execute_PlayerHasNoCurrentCommand_InputIsCommand_CommandFinished(t *testing.T) {
	// Arrange
	conn := &FakeTelnetConnection{}
	q := NewInputQueue(10, 10, logging.NewNullLogger())
	player := absmachine.NewPlayer()
	world := absmachine.NewWorld()
	fakeCommand := FakeCommand{
		returnError:  nil,
		returnResult: mudio.CommandResult{},
	}

	world.AddPlayers([]*absmachine.Player{player})

	q.Append(&PlayerInput{
		player:             player,
		command:            &fakeCommand,
		connection:         conn,
		errorReturnChannel: make(chan<- error, 1),
	})

	// Act
	q.Execute(world)

	// Assert
	if fakeCommand.receivedContext == nil {
		t.Error("Fake command was never called?")
	}

	if q.playerQueues[player].currentCommand != nil {
		t.Error("Expected current command for player to be cleared!")
	}

	if conn.writtenText != "[H:0] [M:0] > " {
		t.Errorf("Unexpected output sent to player Connection: %v", conn.writtenText)
	}
}

func Test_Execute_PlayerHasNoCurrentCommand_InputIsCommand_CommandErrorsAreWrittenToConnection(t *testing.T) {
	// Arrange
	conn := &FakeTelnetConnection{}
	q := NewInputQueue(10, 10, logging.NewNullLogger())
	player := absmachine.NewPlayer()
	world := absmachine.NewWorld()
	fakeCommand := FakeCommand{
		returnError:  mudio.NewCommandError("foo"),
		returnResult: mudio.CommandResult{},
	}

	world.AddPlayers([]*absmachine.Player{player})

	q.Append(&PlayerInput{
		player:             player,
		command:            &fakeCommand,
		connection:         conn,
		errorReturnChannel: make(chan<- error, 1),
	})

	// Act
	q.Execute(world)

	// Assert
	if fakeCommand.receivedContext == nil {
		t.Error("Fake command was never called?")
	}

	if q.playerQueues[player].currentCommand != nil {
		t.Error("Expected current command for player to be cleared!")
	}

	if conn.writtenText != "foo\r\n[H:0] [M:0] > " {
		t.Errorf("Unexpected output sent to player Connection: %v", conn.writtenText)
	}
}

func Test_Execute_PlayerHasNoCurrentCommand_InputIsCommand_CommandWantsToContinue(t *testing.T) {
	// Arrange
	conn := &FakeTelnetConnection{}
	q := NewInputQueue(10, 10, logging.NewNullLogger())
	player := absmachine.NewPlayer()
	world := absmachine.NewWorld()
	fakeCommand := FakeCommand{
		returnError:  nil,
		returnResult: mudio.ContinueWithPrompt("A prompt"),
	}

	world.AddPlayers([]*absmachine.Player{player})

	q.Append(&PlayerInput{
		player:             player,
		command:            &fakeCommand,
		connection:         conn,
		errorReturnChannel: make(chan<- error, 1),
	})

	// Act
	q.Execute(world)

	// Assert
	if fakeCommand.receivedContext == nil {
		t.Error("Fake command was never called?")
	}

	if q.playerQueues[player].currentCommand != &fakeCommand {
		t.Error("Expected current command for player to remain!")
	}

	if conn.writtenText != "A prompt" {
		t.Errorf("Unexpected output sent to player Connection: %v", conn.writtenText)
	}
}

func Test_Execute_PlayerHasNoCurrentCommand_InputIsCommand_CommandWantsToTerminate(t *testing.T) {
	// Arrange
	conn := &FakeTelnetConnection{}
	q := NewInputQueue(10, 10, logging.NewNullLogger())
	player := absmachine.NewPlayer()
	world := absmachine.NewWorld()
	fakeCommand := FakeCommand{
		returnError: nil,
		returnResult: mudio.CommandResult{
			TerminatationRequested: true,
		},
	}
	errorReturnChannel := make(chan error, 10)

	q.commandParser = func(text string) (command mudio.Command, err error) {
		return &fakeCommand, nil
	}

	world.AddPlayers([]*absmachine.Player{player})

	q.Append(&PlayerInput{
		player:             player,
		command:            &fakeCommand,
		connection:         conn,
		errorReturnChannel: errorReturnChannel,
	})

	// Act
	q.Execute(world)

	// Assert
	if fakeCommand.receivedContext == nil {
		t.Error("Fake command was never called?")
	}

	if q.playerQueues[player].currentCommand != nil {
		t.Error("Expected current command for player to be cleared!")
	}

	select {
	case err := <-errorReturnChannel:
		if err != ErrPlayerQuit {
			t.Errorf("Did not expect '%v' on the error return channel", err)
		}
	default:
		t.Error("Expected an error on the error return channel!")
	}
}
*/
