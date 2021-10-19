package mudio

import (
	"fmt"

	"github.com/jorgensigvardsson/gomud/absmachine"
)

/**** Command: Login ****/
type LoginState int

const (
	LS_Initial LoginState = iota
	LS_WantUsername
	LS_WantPassword
)

type CommandLogin struct {
	username string
	state    LoginState
}

func NewCommandLogin(args []string) (Command, CommandRequirementsEvaluator) {
	return &CommandLogin{state: LS_Initial}, nil
}

func (command *CommandLogin) Execute(context *CommandContext) (CommandResult, *CommandError) {
	switch command.state {
	case LS_Initial:
		// Show message of the day to user and set command's state to LS_WantUsername
		command.state = LS_WantUsername
		return CommandResult{Prompt: "$fg_bcyan$Username: ", Output: "Welcome to GO mud!\r\n" /* TODO: Read from file */}, nil
	case LS_WantUsername:
		command.username = context.Input
		command.state = LS_WantPassword
		return CommandResult{Prompt: "$fg_bcyan$Password: ", TurnOffEcho: true}, nil
	case LS_WantPassword:
		if context.World.HasPlayer(command.username) {
			return CommandResult{
				TerminatationRequested: true,
				TurnOnEcho:             true,
				Output:                 "\r\n", /* Because echo off "stole" the new line from the user */
			}, &CommandError{"You are already logged in from another computer."}
		}
		context.Player.Name = command.username
		context.Player.State.SetFlag(absmachine.PS_LOGGED_IN)
		context.World.AddPlayers([]*absmachine.Player{context.Player})
		context.Player.RelocateToRoom(context.World.StartRoom)

		lookResult, _ := lookRoom(context)

		return CommandResult{
			Output:     "\n" + /* Because echo off "stole" the new line from the user */ lookResult.Output,
			TurnOnEcho: true,
		}, nil
	default:
		return CommandResult{TerminatationRequested: true}, &CommandError{fmt.Sprintf("Unknown state reached: %v, preventing player from logging in.", command.state)}
	}
}
