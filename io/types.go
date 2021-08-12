package io

import (
	"fmt"

	"github.com/jorgensigvardsson/gomud/absmachine"
	"github.com/jorgensigvardsson/gomud/mudio"
)

type PlayerEvent int

const (
	PE_Nothing PlayerEvent = iota
	PE_Exited
	PE_EventCount
)

type EchoState int

const (
	ES_None EchoState = iota
	ES_On
	ES_Off
)

type PlayerInput struct {
	player             *absmachine.Player
	text               string
	command            mudio.Command
	outputChannel      chan<- *PlayerOutput
	errorReturnChannel chan<- error
	event              PlayerEvent
}

type PlayerOutput struct {
	text      string
	echoState EchoState
}

func NewCommandPlayerInput(command mudio.Command, player *absmachine.Player, errorReturnChannel chan<- error, outputChannel chan<- *PlayerOutput) *PlayerInput {
	return &PlayerInput{
		player:             player,
		command:            command,
		errorReturnChannel: errorReturnChannel,
		outputChannel:      outputChannel,
	}
}

func NewEventPlayerInput(event PlayerEvent, player *absmachine.Player, errorReturnChannel chan<- error, outputChannel chan<- *PlayerOutput) *PlayerInput {
	return &PlayerInput{
		player:             player,
		event:              event,
		errorReturnChannel: errorReturnChannel,
		outputChannel:      outputChannel,
	}
}

func NewTextPlayerInput(text string, player *absmachine.Player, errorReturnChannel chan<- error, outputChannel chan<- *PlayerOutput) *PlayerInput {
	return &PlayerInput{
		player:             player,
		text:               text,
		errorReturnChannel: errorReturnChannel,
		outputChannel:      outputChannel,
	}
}

func PrintlnOutput(args ...interface{}) *PlayerOutput {
	return &PlayerOutput{
		text: fmt.Sprintln(args...),
	}
}

func PrintlnfOutput(text string, args ...interface{}) *PlayerOutput {
	return &PlayerOutput{
		text: fmt.Sprintln(fmt.Sprintf(text, args...)),
	}
}

func PrintfOutput(text string, args ...interface{}) *PlayerOutput {
	return &PlayerOutput{
		text: fmt.Sprintf(text, args...),
	}
}

func PrintOutput(text string) *PlayerOutput {
	return &PlayerOutput{
		text: text,
	}
}

func EchoOnOutput() *PlayerOutput {
	return &PlayerOutput{
		echoState: ES_On,
	}
}

func EchoOffOutput() *PlayerOutput {
	return &PlayerOutput{
		echoState: ES_Off,
	}
}
