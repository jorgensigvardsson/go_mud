package main

import (
	"container/list"

	"github.com/jorgensigvardsson/gomud/absmachine"
	"github.com/jorgensigvardsson/gomud/mudio"
)

type PlayerInputOrCommand struct {
	connection mudio.TelnetConnection
	player     *absmachine.Player
	input      string
	command    mudio.Command
}

type QueueEntry struct {
	tick           uint64
	inputOrCommand *PlayerInputOrCommand
}

type InputQueue struct {
	currentTick uint64
	inputs      *list.List
}

func NewInputQueue() *InputQueue {
	return &InputQueue{
		inputs: list.New(),
	}
}

type InputHandler func(playerInput *PlayerInputOrCommand)

func (q *InputQueue) ForEachCurrentTick(handler InputHandler) {
	iterator := q.inputs.Front()

	if iterator == nil {
		return
	}

	playersProcessed := make(map[*absmachine.Player]bool)
	for iterator != nil {
		queuedPlayerInputOrCommand := iterator.Value.(*QueueEntry)

		// Is the input from before the current tick, or on this tick? If so, then consider it!
		if queuedPlayerInputOrCommand.tick <= q.currentTick {
			// Has this player already been processed in this loop?
			_, isProcessedAlready := playersProcessed[queuedPlayerInputOrCommand.inputOrCommand.player]
			if isProcessedAlready {
				// Move it forward in time so that this input is processed in the next tick
				queuedPlayerInputOrCommand.tick = q.currentTick + 1

				// Nothing to do, move on!
				iterator = iterator.Next()
			} else {
				handler(queuedPlayerInputOrCommand.inputOrCommand)

				// Mark as processed already (to make sure no other queued message is touched, even if we didn't process any input!)
				playersProcessed[queuedPlayerInputOrCommand.inputOrCommand.player] = true

				// Unlink this input and move on to next
				tempIterator := iterator.Next()
				q.inputs.Remove(iterator)
				iterator = tempIterator
			}
		} else {
			// Nothing to do, check next!
			iterator = iterator.Next()
		}
	}
}

func (q *InputQueue) Append(inputOrCommand *PlayerInputOrCommand) {
	q.inputs.PushBack(&QueueEntry{
		tick:           q.currentTick,
		inputOrCommand: inputOrCommand,
	})
}

func (q *InputQueue) Tick() {
	q.currentTick++
}
