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

func NewCommandLogin(args []string) Command {
	return &CommandLogin{state: LS_Initial}
}

func (command *CommandLogin) Execute(context *CommandContext) (CommandResult, error) {
	switch command.state {
	case LS_Initial:
		// Show message of the day to user
		showMotd(context.Connection)
		command.state = LS_WantUsername
		return ContinueWithPrompt("Username: "), nil
	case LS_WantUsername:
		command.username = context.Input
		command.state = LS_WantPassword
		context.Connection.EchoOff()
		return ContinueWithPrompt("Password: "), nil
	case LS_WantPassword:
		context.Connection.EchoOn()
		context.Connection.WriteLine("") // Emit new line, because echo off "stole it" when the user entered password

		if context.World.HasPlayer(command.username) {
			return CommandResult{TerminatationRequested: true}, &CommandError{"You are already logged in from another computer."}
		}
		context.Player.Name = command.username
		context.Player.State.SetFlag(absmachine.PS_LOGGED_IN)
		context.World.AddPlayers([]*absmachine.Player{context.Player})
		return CommandFinished, nil
	default:
		return CommandResult{TerminatationRequested: true}, &CommandError{fmt.Sprintf("Unknown state reached: %v, preventing player from logging in.", command.state)}
	}
}

func showMotd(telnetConnection TelnetConnection) {
	// TODO: Read MOTD from file
	telnetConnection.WriteLine("Welcome to GO mud!")
}
