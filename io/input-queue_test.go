package io

import (
	"errors"
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

	if promptText != "$fg_bcyan$[H:103] [M:43] > " {
		t.Errorf("Unexpected prompt: %v", promptText)
	}
}

func Test_Append_PanicsIfNoOutputChannel(t *testing.T) {
	// Arrange
	q := NewInputQueue(1, 1, logging.NewNullLogger())
	p := absmachine.NewPlayer()

	errorChannel := make(chan error)

	// Act && Assert
	defer func() {
		if recover() == nil {
			t.Error("Append did not panic as expected!")
		}
	}()

	q.Append(
		&PlayerInput{
			player:             p,
			text:               "cmd",
			errorReturnChannel: errorChannel,
		},
	)
}

func Test_Append_PanicsIfNoErrorReturnChannel(t *testing.T) {
	// Arrange
	q := NewInputQueue(1, 1, logging.NewNullLogger())
	p := absmachine.NewPlayer()

	outputChannel := make(chan *PlayerOutput)

	// Act && Assert
	defer func() {
		if recover() == nil {
			t.Error("Append did not panic as expected!")
		}
	}()

	q.Append(
		&PlayerInput{
			player:        p,
			text:          "cmd",
			outputChannel: outputChannel,
		},
	)
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

	testError(t, errorChannel1)
	testError(t, errorChannel2, ErrTooManyPlayers)
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

	testError(t, errorChannel, ErrTooMuchInput)
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
	testOutput(t, outputChannel, "$fg_bcyan$[H:0] [M:0] > ")
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

	testOutput(t, outputChannel, fmt.Sprintln("$fg_bred$foo"), "$fg_bcyan$[H:0] [M:0] > ")
}

func Test_Execute_PlayerHasCurrentCommand_InputIsSentToCurrentCommand_CommandWantsToContinue(t *testing.T) {
	// Arrange
	q := NewInputQueue(10, 10, logging.NewNullLogger())
	player := absmachine.NewPlayer()
	world := absmachine.NewWorld()
	outputChannel := make(chan *PlayerOutput, 10)
	currentCommand := FakeCommand{
		returnError:  nil,
		returnResult: mudio.CommandResult{Prompt: "A prompt"},
	}

	world.AddPlayers([]*absmachine.Player{player})
	q.playerQueues[player] = newPlayerQueue()
	q.playerQueues[player].currentCommand = &currentCommand

	q.Append(&PlayerInput{
		player:             player,
		text:               "cmd text",
		outputChannel:      outputChannel,
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

	testOutput(t, outputChannel, "A prompt")
}

func Test_Execute_PlayerHasCurrentCommand_InputIsSentToCurrentCommand_CommandWantsToTerminate(t *testing.T) {
	// Arrange
	q := NewInputQueue(10, 10, logging.NewNullLogger())
	player := absmachine.NewPlayer()
	world := absmachine.NewWorld()
	outputChannel := make(chan *PlayerOutput, 10)
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
		outputChannel:      outputChannel,
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

	testError(t, errorReturnChannel, ErrPlayerQuit)
}

func Test_Execute_PlayerHasNoCurrentCommand_InvalidInput_ErrorAndStandardPromptWrittenToConnection(t *testing.T) {
	// Arrange
	q := NewInputQueue(10, 10, logging.NewNullLogger())
	player := absmachine.NewPlayer()
	world := absmachine.NewWorld()
	outputChannel := make(chan *PlayerOutput, 10)

	q.commandParser = func(text string, player *absmachine.Player) (command mudio.Command, err error) {
		return nil, errors.New("foo")
	}

	world.AddPlayers([]*absmachine.Player{player})

	q.Append(&PlayerInput{
		player:             player,
		text:               "gargksjdl",
		outputChannel:      outputChannel,
		errorReturnChannel: make(chan<- error, 1),
	})

	// Act
	q.Execute(world)

	// Assert
	testOutput(t, outputChannel, fmt.Sprintln("$fg_bred$foo"), "$fg_bcyan$[H:0] [M:0] > ")
}

func Test_Execute_PlayerHasNoCurrentCommand_InputIsSentToParsedCommand_CommandFinished(t *testing.T) {
	// Arrange
	q := NewInputQueue(10, 10, logging.NewNullLogger())
	player := absmachine.NewPlayer()
	world := absmachine.NewWorld()
	outputChannel := make(chan *PlayerOutput, 10)
	fakeCommand := FakeCommand{
		returnError:  nil,
		returnResult: mudio.CommandResult{},
	}

	q.commandParser = func(text string, player *absmachine.Player) (command mudio.Command, err error) {
		return &fakeCommand, nil
	}

	world.AddPlayers([]*absmachine.Player{player})

	q.Append(&PlayerInput{
		player:             player,
		text:               "cmd text",
		outputChannel:      outputChannel,
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

	testOutput(t, outputChannel, "$fg_bcyan$[H:0] [M:0] > ")
}

func Test_Execute_PlayerHasNoCurrentCommand_InputIsSentToParsedCommand_CommandErrorsAreWrittenToConnection(t *testing.T) {
	// Arrange
	q := NewInputQueue(10, 10, logging.NewNullLogger())
	player := absmachine.NewPlayer()
	world := absmachine.NewWorld()
	outputChannel := make(chan *PlayerOutput, 10)
	fakeCommand := FakeCommand{
		returnError:  mudio.NewCommandError("foo"),
		returnResult: mudio.CommandResult{},
	}

	q.commandParser = func(text string, player *absmachine.Player) (command mudio.Command, err error) {
		return &fakeCommand, nil
	}

	world.AddPlayers([]*absmachine.Player{player})

	q.Append(&PlayerInput{
		player:             player,
		text:               "cmd text",
		outputChannel:      outputChannel,
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

	testOutput(t, outputChannel, fmt.Sprintln("$fg_bred$foo"), "$fg_bcyan$[H:0] [M:0] > ")
}

func Test_Execute_PlayerHasNoCurrentCommand_InputIsSentToParsedCommand_CommandWantsToContinue(t *testing.T) {
	// Arrange
	q := NewInputQueue(10, 10, logging.NewNullLogger())
	player := absmachine.NewPlayer()
	world := absmachine.NewWorld()
	outputChannel := make(chan *PlayerOutput, 10)
	fakeCommand := FakeCommand{
		returnError:  nil,
		returnResult: mudio.CommandResult{Prompt: "A prompt"},
	}

	q.commandParser = func(text string, player *absmachine.Player) (command mudio.Command, err error) {
		return &fakeCommand, nil
	}

	world.AddPlayers([]*absmachine.Player{player})

	q.Append(&PlayerInput{
		player:             player,
		text:               "cmd text",
		outputChannel:      outputChannel,
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

	testOutput(t, outputChannel, "A prompt")
}

func Test_Execute_PlayerHasNoCurrentCommand_InputIsSentToParsedCommand_CommandWantsToTerminate(t *testing.T) {
	// Arrange
	q := NewInputQueue(10, 10, logging.NewNullLogger())
	player := absmachine.NewPlayer()
	world := absmachine.NewWorld()
	outputChannel := make(chan *PlayerOutput, 10)
	fakeCommand := FakeCommand{
		returnError: nil,
		returnResult: mudio.CommandResult{
			TerminatationRequested: true,
		},
	}
	errorReturnChannel := make(chan error, 10)

	q.commandParser = func(text string, player *absmachine.Player) (command mudio.Command, err error) {
		return &fakeCommand, nil
	}

	world.AddPlayers([]*absmachine.Player{player})

	q.Append(&PlayerInput{
		player:             player,
		text:               "cmd text",
		outputChannel:      outputChannel,
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

	testError(t, errorReturnChannel, ErrPlayerQuit)
}

func Test_Execute_PlayerHasNoCurrentCommand_InputIsCommand_CommandFinished(t *testing.T) {
	// Arrange
	q := NewInputQueue(10, 10, logging.NewNullLogger())
	player := absmachine.NewPlayer()
	world := absmachine.NewWorld()
	outputChannel := make(chan *PlayerOutput, 10)
	fakeCommand := FakeCommand{
		returnError:  nil,
		returnResult: mudio.CommandResult{},
	}

	world.AddPlayers([]*absmachine.Player{player})

	q.Append(&PlayerInput{
		player:             player,
		command:            &fakeCommand,
		outputChannel:      outputChannel,
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

	testOutput(t, outputChannel, "$fg_bcyan$[H:0] [M:0] > ")
}

func Test_Execute_PlayerHasNoCurrentCommand_InputIsCommand_CommandErrorsAreWrittenToConnection(t *testing.T) {
	// Arrange
	q := NewInputQueue(10, 10, logging.NewNullLogger())
	player := absmachine.NewPlayer()
	world := absmachine.NewWorld()
	outputChannel := make(chan *PlayerOutput, 10)
	fakeCommand := FakeCommand{
		returnError:  mudio.NewCommandError("foo"),
		returnResult: mudio.CommandResult{},
	}

	world.AddPlayers([]*absmachine.Player{player})

	q.Append(&PlayerInput{
		player:             player,
		command:            &fakeCommand,
		outputChannel:      outputChannel,
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

	testOutput(t, outputChannel, fmt.Sprintln("$fg_bred$foo"), "$fg_bcyan$[H:0] [M:0] > ")
}

func Test_Execute_PlayerHasNoCurrentCommand_InputIsCommand_CommandWantsToContinue(t *testing.T) {
	// Arrange
	q := NewInputQueue(10, 10, logging.NewNullLogger())
	player := absmachine.NewPlayer()
	world := absmachine.NewWorld()
	outputChannel := make(chan *PlayerOutput, 10)
	fakeCommand := FakeCommand{
		returnError:  nil,
		returnResult: mudio.CommandResult{Prompt: "A prompt"},
	}

	world.AddPlayers([]*absmachine.Player{player})

	q.Append(&PlayerInput{
		player:             player,
		command:            &fakeCommand,
		outputChannel:      outputChannel,
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

	testOutput(t, outputChannel, "A prompt")
}

func Test_Execute_PlayerHasNoCurrentCommand_InputIsCommand_CommandWantsToTerminate(t *testing.T) {
	// Arrange
	q := NewInputQueue(10, 10, logging.NewNullLogger())
	player := absmachine.NewPlayer()
	world := absmachine.NewWorld()
	outputChannel := make(chan *PlayerOutput, 10)
	fakeCommand := FakeCommand{
		returnError: nil,
		returnResult: mudio.CommandResult{
			TerminatationRequested: true,
		},
	}
	errorReturnChannel := make(chan error, 10)

	q.commandParser = func(text string, player *absmachine.Player) (command mudio.Command, err error) {
		return &fakeCommand, nil
	}

	world.AddPlayers([]*absmachine.Player{player})

	q.Append(&PlayerInput{
		player:             player,
		command:            &fakeCommand,
		outputChannel:      outputChannel,
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

	testError(t, errorReturnChannel, ErrPlayerQuit)
}

func Test_Execute_TextMessagesAreSentToRecipientPlayers(t *testing.T) {
	// Arrange
	q := NewInputQueue(10, 10, logging.NewNullLogger())
	player := absmachine.NewPlayer()
	player2 := absmachine.NewPlayer()
	player3 := absmachine.NewPlayer()
	world := absmachine.NewWorld()
	playerOutputChannel := make(chan *PlayerOutput, 10)
	player2OutputChannel := make(chan *PlayerOutput, 10)
	player3OutputChannel := make(chan *PlayerOutput, 10)
	playerErrorReturnChannel := make(chan error, 10)

	world.AddPlayers([]*absmachine.Player{player, player2, player3})

	q.playerQueues[player2] = newPlayerQueue()
	q.playerQueues[player3] = newPlayerQueue()

	q.playerQueues[player2].outputChannel = player2OutputChannel
	q.playerQueues[player3].outputChannel = player3OutputChannel

	player2.Health = 123
	player2.Mana = 321
	player3.Health = 456
	player3.Mana = 654

	fakeCommand := FakeCommand{
		returnError: nil,
		returnResult: mudio.CommandResult{
			TextMessages: []mudio.TextMessage{
				{
					Text:            "for player 2",
					RecipientPlayer: player2,
				},
				{
					Text:            "for player 3",
					RecipientPlayer: player3,
				},
			},
		},
	}

	q.Append(&PlayerInput{
		player:             player,
		command:            &fakeCommand,
		outputChannel:      playerOutputChannel,
		errorReturnChannel: playerErrorReturnChannel,
	})

	// Act
	q.Execute(world)

	// Assert
	testOutput(t, player2OutputChannel, fmt.Sprintln(""), fmt.Sprintln("for player 2"), "$fg_bcyan$[H:123] [M:321] > ")
	testOutput(t, player3OutputChannel, fmt.Sprintln(""), fmt.Sprintln("for player 3"), "$fg_bcyan$[H:456] [M:654] > ")
}

func Test_Execute_EchoMaybeTurnedOff(t *testing.T) {
	// Arrange
	q := NewInputQueue(10, 10, logging.NewNullLogger())
	player := absmachine.NewPlayer()
	world := absmachine.NewWorld()
	playerOutputChannel := make(chan *PlayerOutput, 10)
	playerErrorReturnChannel := make(chan error, 10)

	world.AddPlayers([]*absmachine.Player{player})

	fakeCommand := FakeCommand{
		returnError: nil,
		returnResult: mudio.CommandResult{
			TurnOffEcho: true,
		},
	}

	q.Append(&PlayerInput{
		player:             player,
		command:            &fakeCommand,
		outputChannel:      playerOutputChannel,
		errorReturnChannel: playerErrorReturnChannel,
	})

	// Act
	q.Execute(world)

	// Assert
	output := getOutput(playerOutputChannel)
	if len(output) != 2 || output[0].text != "$fg_bcyan$[H:0] [M:0] >" && output[1].echoState != ES_Off {
		t.Error("Expected output to be a prompt and a single ES_Off")
	}
}

func Test_Execute_EchoMaybeTurnedOn(t *testing.T) {
	// Arrange
	q := NewInputQueue(10, 10, logging.NewNullLogger())
	player := absmachine.NewPlayer()
	world := absmachine.NewWorld()
	playerOutputChannel := make(chan *PlayerOutput, 10)
	playerErrorReturnChannel := make(chan error, 10)

	world.AddPlayers([]*absmachine.Player{player})

	fakeCommand := FakeCommand{
		returnError: nil,
		returnResult: mudio.CommandResult{
			TurnOnEcho: true,
		},
	}

	q.Append(&PlayerInput{
		player:             player,
		command:            &fakeCommand,
		outputChannel:      playerOutputChannel,
		errorReturnChannel: playerErrorReturnChannel,
	})

	// Act
	q.Execute(world)

	// Assert
	output := getOutput(playerOutputChannel)
	if len(output) != 2 || output[0].text != "$fg_bcyan$[H:0] [M:0] >" && output[1].echoState != ES_On {
		t.Error("Expected output to be a prompt and a single ES_On")
	}
}

func Test_Execute_CommandOutputSentToPlayer(t *testing.T) {
	// Arrange
	q := NewInputQueue(10, 10, logging.NewNullLogger())
	player := absmachine.NewPlayer()
	world := absmachine.NewWorld()
	playerOutputChannel := make(chan *PlayerOutput, 10)
	playerErrorReturnChannel := make(chan error, 10)

	world.AddPlayers([]*absmachine.Player{player})

	fakeCommand := FakeCommand{
		returnError: nil,
		returnResult: mudio.CommandResult{
			Output: "Some output",
		},
	}

	q.Append(&PlayerInput{
		player:             player,
		command:            &fakeCommand,
		outputChannel:      playerOutputChannel,
		errorReturnChannel: playerErrorReturnChannel,
	})

	// Act
	q.Execute(world)

	// Assert
	testOutput(t, playerOutputChannel, fmt.Sprintln("Some output"), "$fg_bcyan$[H:0] [M:0] > ")
}

// Utilities for testing the input queue
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

func getErrors(channel <-chan error) []error {
	errors := make([]error, 0)
	done := false

	for !done {
		select {
		case err := <-channel:
			errors = append(errors, err)
		default:
			done = true
		}
	}

	return errors
}

func testOutput(t *testing.T, outputChannel <-chan *PlayerOutput, expectedValues ...string) {
	output := getOutput(outputChannel)

	isError := false
	if len(output) != len(expectedValues) {
		isError = true
	} else {
		for i := 0; !isError && i < len(output); i++ {
			if output[i].text != expectedValues[i] {
				isError = true
			}
		}
	}

	if isError {
		outputText := make([]string, len(output))
		for i := 0; i < len(output); i++ {
			outputText[i] = output[i].text
		}
		t.Errorf("Expected output: %#v, but got: %#v", expectedValues, outputText)
	}
}

func testError(t *testing.T, errorReturnChannel <-chan error, expectedErrors ...error) {
	errors := getErrors(errorReturnChannel)

	isError := false
	if len(errors) != len(expectedErrors) {
		isError = true
	} else {
		for i := 0; !isError && i < len(errors); i++ {
			if errors[i] != expectedErrors[i] {
				isError = true
			}
		}
	}

	if isError {
		actualErrorTexts := make([]string, len(errors))
		for i := 0; i < len(errors); i++ {
			actualErrorTexts[i] = errors[i].Error()
		}

		expectedErrorTexts := make([]string, len(errors))
		for i := 0; i < len(expectedErrors); i++ {
			expectedErrorTexts[i] = expectedErrors[i].Error()
		}
		t.Errorf("Expected errors: %#v, but got: %#v", expectedErrorTexts, actualErrorTexts)
	}
}
