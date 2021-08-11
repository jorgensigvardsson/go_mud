package main

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jorgensigvardsson/gomud/absmachine"
	"github.com/jorgensigvardsson/gomud/mudio"
)

func Test_showNormalPrompt(t *testing.T) {
	conn := &FakeTelnetConnection{}
	player := absmachine.Player{
		Health: 103,
		Mana:   43,
	}

	showNormalPrompt(conn, &player)

	if conn.writtenText != "[H:103] [M:43] > " {
		t.Errorf("Unexpected prompt: %v", conn.writtenText)
	}
}

type FakeTelnetConnection struct {
	writtenText string
}

func (conn *FakeTelnetConnection) ReadLine() (line string, err error) {
	panic("ReadLine not implemented")
}

func (conn *FakeTelnetConnection) WriteLine(line string) error {
	conn.writtenText += line + "\r\n"
	return nil
}

func (conn *FakeTelnetConnection) WriteLinef(line string, args ...interface{}) error {
	conn.writtenText += fmt.Sprintf(line, args...) + "\r\n"
	return nil
}

func (conn *FakeTelnetConnection) WriteString(text string) error {
	conn.writtenText += text
	return nil
}

func (conn *FakeTelnetConnection) WriteStringf(text string, args ...interface{}) error {
	conn.writtenText += fmt.Sprintf(text, args...)
	return nil
}

func (conn *FakeTelnetConnection) EchoOff() error { panic("EchoOff not implemented") }
func (conn *FakeTelnetConnection) EchoOn() error  { panic("EchoOn not implemented") }
func (conn *FakeTelnetConnection) Close() error   { panic("Close not implemented") }

func Test_Append_PlayerLimitIsRespected(t *testing.T) {
	q := NewInputQueue(1, 1)
	p1 := absmachine.NewPlayer()
	p2 := absmachine.NewPlayer()
	errorChannel1 := make(chan error, 10)
	errorChannel2 := make(chan error, 10)

	p1.Name = "p1"
	p2.Name = "p2"

	q.Append(
		&PlayerInput{
			player:             p1,
			text:               "cmd",
			errorReturnChannel: errorChannel1,
		},
	)

	q.Append(
		&PlayerInput{
			player:             p2,
			text:               "cmd",
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
	q := NewInputQueue(1, 1)
	p := absmachine.NewPlayer()
	errorChannel := make(chan error, 10)

	p.Name = "p1"

	q.Append(
		&PlayerInput{
			player:             p,
			text:               "cmd 1",
			errorReturnChannel: errorChannel,
		},
	)

	q.Append(
		&PlayerInput{
			player:             p,
			text:               "cmd 2",
			errorReturnChannel: errorChannel,
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
	q := NewInputQueue(10, 10)
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
	q := NewInputQueue(10, 10)
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
	q := NewInputQueue(10, 10)
	player := absmachine.NewPlayer()
	world := absmachine.NewWorld()

	world.AddPlayers([]*absmachine.Player{player})
	q.Append(&PlayerInput{
		player: player,
		event:  PE_Exited,
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
	q := NewInputQueue(10, 10)
	player := absmachine.NewPlayer()
	world := absmachine.NewWorld()

	world.AddPlayers([]*absmachine.Player{player})
	q.Append(&PlayerInput{
		player: player,
		event:  PE_EventCount + 1,
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
	returnError     error
	returnResult    mudio.CommandResult
}

func (cmd *FakeCommand) Execute(context *mudio.CommandContext) (result mudio.CommandResult, err error) {
	cmd.receivedContext = context
	return cmd.returnResult, cmd.returnError
}

func Test_Execute_NoInput_StandardPromptWrittenToConnection(t *testing.T) {
	// Arrange
	conn := &FakeTelnetConnection{}
	q := NewInputQueue(10, 10)
	player := absmachine.NewPlayer()
	world := absmachine.NewWorld()

	world.AddPlayers([]*absmachine.Player{player})

	q.Append(&PlayerInput{
		player:     player,
		connection: conn,
	})

	// Act
	q.Execute(world)

	// Assert
	if conn.writtenText != "[H:0] [M:0] > " {
		t.Errorf("Unexpected output sent to player connection: %v", conn.writtenText)
	}
}

func Test_Execute_PlayerHasCurrentCommand_InputIsSentToCurrentCommand_CommandFinished(t *testing.T) {
	// Arrange
	conn := &FakeTelnetConnection{}
	q := NewInputQueue(10, 10)
	player := absmachine.NewPlayer()
	world := absmachine.NewWorld()
	currentCommand := FakeCommand{
		returnError:  nil,
		returnResult: mudio.CommandFinished,
	}

	world.AddPlayers([]*absmachine.Player{player})
	q.playerQueues[player] = newPlayerQueue()
	q.playerQueues[player].currentCommand = &currentCommand

	q.Append(&PlayerInput{
		player:     player,
		text:       "cmd text",
		connection: conn,
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
	conn := &FakeTelnetConnection{}
	q := NewInputQueue(10, 10)
	player := absmachine.NewPlayer()
	world := absmachine.NewWorld()
	currentCommand := FakeCommand{
		returnError:  errors.New("foo"),
		returnResult: mudio.CommandFinished,
	}

	world.AddPlayers([]*absmachine.Player{player})
	q.playerQueues[player] = newPlayerQueue()
	q.playerQueues[player].currentCommand = &currentCommand

	q.Append(&PlayerInput{
		player:     player,
		text:       "cmd text",
		connection: conn,
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

	if conn.writtenText != "foo\r\n[H:0] [M:0] > " {
		t.Errorf("Unexpected output sent to player connection: %v", conn.writtenText)
	}
}

func Test_Execute_PlayerHasCurrentCommand_InputIsSentToCurrentCommand_CommandWantsToContinue(t *testing.T) {
	// Arrange
	conn := &FakeTelnetConnection{}
	q := NewInputQueue(10, 10)
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
		player:     player,
		text:       "cmd text",
		connection: conn,
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
		t.Errorf("Unexpected output sent to player connection: %v", conn.writtenText)
	}
}

func Test_Execute_PlayerHasCurrentCommand_InputIsSentToCurrentCommand_CommandWantsToTerminate(t *testing.T) {
	// Arrange
	conn := &FakeTelnetConnection{}
	q := NewInputQueue(10, 10)
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
	q := NewInputQueue(10, 10)
	player := absmachine.NewPlayer()
	world := absmachine.NewWorld()

	q.commandParser = func(text string) (command mudio.Command, err error) {
		return nil, errors.New("foo")
	}

	world.AddPlayers([]*absmachine.Player{player})

	q.Append(&PlayerInput{
		player:     player,
		text:       "gargksjdl",
		connection: conn,
	})

	// Act
	q.Execute(world)

	// Assert
	if conn.writtenText != "foo\r\n[H:0] [M:0] > " {
		t.Errorf("Unexpected output sent to player connection: %v", conn.writtenText)
	}
}

func Test_Execute_PlayerHasNoCurrentCommand_InputIsSentToParsedCommand_CommandFinished(t *testing.T) {
	// Arrange
	conn := &FakeTelnetConnection{}
	q := NewInputQueue(10, 10)
	player := absmachine.NewPlayer()
	world := absmachine.NewWorld()
	fakeCommand := FakeCommand{
		returnError:  nil,
		returnResult: mudio.CommandFinished,
	}

	q.commandParser = func(text string) (command mudio.Command, err error) {
		return &fakeCommand, nil
	}

	world.AddPlayers([]*absmachine.Player{player})

	q.Append(&PlayerInput{
		player:     player,
		text:       "cmd text",
		connection: conn,
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
		t.Errorf("Unexpected output sent to player connection: %v", conn.writtenText)
	}
}

func Test_Execute_PlayerHasNoCurrentCommand_InputIsSentToParsedCommand_CommandErrorsAreWrittenToConnection(t *testing.T) {
	// Arrange
	conn := &FakeTelnetConnection{}
	q := NewInputQueue(10, 10)
	player := absmachine.NewPlayer()
	world := absmachine.NewWorld()
	fakeCommand := FakeCommand{
		returnError:  errors.New("foo"),
		returnResult: mudio.CommandFinished,
	}

	q.commandParser = func(text string) (command mudio.Command, err error) {
		return &fakeCommand, nil
	}

	world.AddPlayers([]*absmachine.Player{player})

	q.Append(&PlayerInput{
		player:     player,
		text:       "cmd text",
		connection: conn,
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
		t.Errorf("Unexpected output sent to player connection: %v", conn.writtenText)
	}
}

func Test_Execute_PlayerHasNoCurrentCommand_InputIsSentToParsedCommand_CommandWantsToContinue(t *testing.T) {
	// Arrange
	conn := &FakeTelnetConnection{}
	q := NewInputQueue(10, 10)
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
		player:     player,
		text:       "cmd text",
		connection: conn,
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
		t.Errorf("Unexpected output sent to player connection: %v", conn.writtenText)
	}
}

func Test_Execute_PlayerHasNoCurrentCommand_InputIsSentToParsedCommand_CommandWantsToTerminate(t *testing.T) {
	// Arrange
	conn := &FakeTelnetConnection{}
	q := NewInputQueue(10, 10)
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
	q := NewInputQueue(10, 10)
	player := absmachine.NewPlayer()
	world := absmachine.NewWorld()
	fakeCommand := FakeCommand{
		returnError:  nil,
		returnResult: mudio.CommandFinished,
	}

	world.AddPlayers([]*absmachine.Player{player})

	q.Append(&PlayerInput{
		player:     player,
		command:    &fakeCommand,
		connection: conn,
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
		t.Errorf("Unexpected output sent to player connection: %v", conn.writtenText)
	}
}

func Test_Execute_PlayerHasNoCurrentCommand_InputIsCommand_CommandErrorsAreWrittenToConnection(t *testing.T) {
	// Arrange
	conn := &FakeTelnetConnection{}
	q := NewInputQueue(10, 10)
	player := absmachine.NewPlayer()
	world := absmachine.NewWorld()
	fakeCommand := FakeCommand{
		returnError:  errors.New("foo"),
		returnResult: mudio.CommandFinished,
	}

	world.AddPlayers([]*absmachine.Player{player})

	q.Append(&PlayerInput{
		player:     player,
		command:    &fakeCommand,
		connection: conn,
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
		t.Errorf("Unexpected output sent to player connection: %v", conn.writtenText)
	}
}

func Test_Execute_PlayerHasNoCurrentCommand_InputIsCommand_CommandWantsToContinue(t *testing.T) {
	// Arrange
	conn := &FakeTelnetConnection{}
	q := NewInputQueue(10, 10)
	player := absmachine.NewPlayer()
	world := absmachine.NewWorld()
	fakeCommand := FakeCommand{
		returnError:  nil,
		returnResult: mudio.ContinueWithPrompt("A prompt"),
	}

	world.AddPlayers([]*absmachine.Player{player})

	q.Append(&PlayerInput{
		player:     player,
		command:    &fakeCommand,
		connection: conn,
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
		t.Errorf("Unexpected output sent to player connection: %v", conn.writtenText)
	}
}

func Test_Execute_PlayerHasNoCurrentCommand_InputIsCommand_CommandWantsToTerminate(t *testing.T) {
	// Arrange
	conn := &FakeTelnetConnection{}
	q := NewInputQueue(10, 10)
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
