package main

import (
	"container/list"
	"errors"

	"github.com/jorgensigvardsson/gomud/absmachine"
	"github.com/jorgensigvardsson/gomud/logging"
	"github.com/jorgensigvardsson/gomud/mudio"
)

type PlayerEvent int

const (
	PE_Nothing PlayerEvent = iota
	PE_Exited
	PE_EventCount
)

type PlayerInput struct {
	connection         mudio.TelnetConnection
	player             *absmachine.Player
	text               string
	command            mudio.Command
	errorReturnChannel chan<- error
	event              PlayerEvent
}

type PlayerQueue struct {
	inputs             *list.List
	currentCommand     mudio.Command
	errorReturnChannel chan<- error
	connection         mudio.TelnetConnection
}

func newPlayerQueue() *PlayerQueue {
	return &PlayerQueue{
		inputs: list.New(),
	}
}

type InputQueue struct {
	commandParser            mudio.CommandParser
	playerQueues             map[*absmachine.Player]*PlayerQueue
	maxPlayerLimit           int
	maxPlayerInputQueueLimit int
	logger                   logging.Logger
}

func NewInputQueue(maxPlayerLimit int, maxPlayerInputQueueLimit int, logger logging.Logger) *InputQueue {
	return &InputQueue{
		commandParser:            mudio.ParseCommand,
		playerQueues:             make(map[*absmachine.Player]*PlayerQueue),
		maxPlayerLimit:           maxPlayerLimit,
		maxPlayerInputQueueLimit: maxPlayerInputQueueLimit,
		logger:                   logger,
	}
}

func (q *InputQueue) Execute(world *absmachine.World) {
	for player, pq := range q.playerQueues {
		if pq.inputs.Len() == 0 {
			continue
		}

		input := pq.inputs.Front().Value.(*PlayerInput)
		pq.inputs.Remove(pq.inputs.Front())

		if input.event != PE_Nothing {
			// If it's an event (rather than input/command),
			// then handle it and go on with the next player queue
			q.handleEvent(input)
			continue
		}

		var command mudio.Command

		if pq.currentCommand != nil {
			command = pq.currentCommand
		} else if input.command != nil {
			command = input.command
		} else if input.text != "" {
			var err error
			command, err = q.commandParser(input.text)

			if err != nil {
				pq.connection.WriteLinef("Error: %v", err.Error())

				// Player typed in something that was not recognized as a command, so just show a prompt and continue
				showNormalPrompt(pq.connection, player)
				continue
			}
		} else {
			// Show the prompt and continue
			showNormalPrompt(pq.connection, player)
			continue
		}

		commandContext := mudio.CommandContext{
			World:      world,
			Player:     player,
			Connection: pq.connection,
			Input:      input.text,
		}

		result, err := command.Execute(&commandContext)

		if err != nil {
			// We had an error, so let's show that to the user!
			pq.connection.WriteLine(err.Error())
		}

		if result.TerminatationRequested {
			// Termination requested! Let's pass it off to the input handling routine
			pq.errorReturnChannel <- ErrPlayerQuit
			// We're done here, so let's make sure the current command is done
			pq.currentCommand = nil
		} else {
			// Command wants to show a prompt? Then do it
			if result.Prompt != "" {
				pq.connection.WriteString(result.Prompt)
			}

			if result.Continue {
				// Command wants to continue execution, so let's save it for the next inputs
				pq.currentCommand = command
				player.State.SetFlag(absmachine.PS_BUSY) // If a command wants to continue executing, then the player is busy
			} else {
				// We're done here, so let's make sure the current command is done
				pq.currentCommand = nil
				showNormalPrompt(pq.connection, player)
				player.State.ClearFlag(absmachine.PS_BUSY) // If the command is complete, then the player is no longer busy
			}

			for _, response := range result.Responses {
				pq, found := q.playerQueues[response.Player]

				if !found {
					q.logger.Printlnf("Tried to send response to player %v from player %v, but receiving player does not have a queue!", response.Player.Name, player.Name)
				} else {
					pq.connection.WriteLine("")                      // Emit a new line in order to clear the prompt on screen
					pq.connection.WriteLine(response.Text)           // Then the response text
					showNormalPrompt(pq.connection, response.Player) // And finally show the prompt again
				}
			}
		}
	}
}

func (q *InputQueue) handleEvent(input *PlayerInput) {
	switch input.event {
	case PE_Exited:
		// Player exited, so remove it from the world
		absmachine.DestroyPlayer(input.player)
		delete(q.playerQueues, input.player)
	}
}

var ErrPlayerQuit = errors.New("player quit")
var ErrTooManyPlayers = errors.New("too many players connected")
var ErrTooMuchInput = errors.New("too many players connected")
var ErrMissingCommInformation = errors.New("no communication information")

func (q *InputQueue) Append(inputOrCommand *PlayerInput) {
	if inputOrCommand.connection == nil || inputOrCommand.errorReturnChannel == nil {
		panic("No error return channel or connection passed onto Append!")
	}

	pq, ok := q.playerQueues[inputOrCommand.player]
	if !ok {
		if len(q.playerQueues)+1 > q.maxPlayerLimit { // Would adding one player queue go above the limit?
			inputOrCommand.errorReturnChannel <- ErrTooManyPlayers // Signal I/O routine
			return
		}

		pq = newPlayerQueue()
		q.playerQueues[inputOrCommand.player] = pq
	}

	if pq.inputs.Len()+1 > q.maxPlayerInputQueueLimit { // Would adding one more input go above the limit?
		inputOrCommand.errorReturnChannel <- ErrTooMuchInput
		return
	}

	// Make sure we remember the communication channels!
	pq.connection = inputOrCommand.connection
	pq.errorReturnChannel = inputOrCommand.errorReturnChannel

	pq.inputs.PushBack(inputOrCommand)
}

func showNormalPrompt(connection mudio.TelnetConnection, player *absmachine.Player) {
	connection.WriteStringf("[H:%v] [M:%v] > ", player.Health, player.Mana)
}
