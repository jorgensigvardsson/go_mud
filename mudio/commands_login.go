package mudio

import "github.com/jorgensigvardsson/gomud/absmachine"

/**** Command: Login ****/
type LoginState int

const (
	LOGIN_STATE_USERNAME LoginState = iota
	LOGIN_STATE_PASSWORD
)

type CommandLogin struct {
	username string
	state    LoginState
}

func NewCommandLogin() Command {
	return &CommandLogin{state: LOGIN_STATE_USERNAME}
}

func (command *CommandLogin) Execute(context *CommandContext) (CommandSubPrompter, error) {
	return command, nil
}

func (command *CommandLogin) Prompt(context *CommandContext) (string, error) {
	switch command.state {
	case LOGIN_STATE_USERNAME:
		return "Username: ", nil
	case LOGIN_STATE_PASSWORD:
		return "Password: ", nil
	default:
		context.TerminationRequested = true
		return "", &CommandError{"Unknown error occurred, preventing you from logging in."}
	}
}

func (command *CommandLogin) ExecuteSubprompt(input string, context *CommandContext) (CommandSubPrompter, error) {
	switch command.state {
	case LOGIN_STATE_USERNAME:
		command.username = input
		command.state = LOGIN_STATE_PASSWORD
		context.Connection.EchoOff()
		return command, nil
	case LOGIN_STATE_PASSWORD:
		// TODO: Validate username and password!
		context.Player.Name = command.username
		context.Player.State.SetFlag(absmachine.PS_LOGGED_IN)
		context.World.AddPlayers([]*absmachine.Player{context.Player})
		context.Connection.EchoOn()
		context.Connection.WriteLine("") // Emit new line, because echo off "stole it" when the user entered password
		return nil, nil
	default:
		context.TerminationRequested = true
		return nil, &CommandError{"Unknown error occurred, preventing you from logging in."}
	}
}
