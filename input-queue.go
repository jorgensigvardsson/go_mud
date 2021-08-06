package main

import (
	"container/list"
	"sync"

	"github.com/jorgensigvardsson/gomud/absmachine"
	"github.com/jorgensigvardsson/gomud/mudio"
)

type PlayerInput struct {
	tick       uint64
	player     *absmachine.Player
	input      string
	connection mudio.TelnetConnection
}

type InputQueue struct {
	lock   sync.Mutex
	inputs *list.List
}

func NewInputQueue() *InputQueue {
	return &InputQueue{
		inputs: list.New(),
	}
}

type InputHandler func(playerInput *PlayerInput)

func (q *InputQueue) ForEachUntilTick(tick uint64, handler InputHandler) {
	q.lock.Lock()
	defer q.lock.Unlock()

	iterator := q.inputs.Front()

	if iterator == nil {
		return
	}

	playersProcessed := make(map[*absmachine.Player]bool)
	for iterator != nil {
		playerInput := iterator.Value.(*PlayerInput)

		// Is the input from before the current tick, or on this tick? If so, then consider it!
		if playerInput.tick <= tick {
			// Has this player already been processed in this loop?
			_, isProcessedAlready := playersProcessed[playerInput.player]
			if isProcessedAlready {
				// Move it forward in time so that this input is processed in the next tick
				playerInput.tick = tick + 1

				// Nothing to do, move on!
				iterator = iterator.Next()
			} else {
				handler(playerInput)

				// Mark as processed already (to make sure no other queued message is touched, even if we didn't process any input!)
				playersProcessed[playerInput.player] = true

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

func (q *InputQueue) Append(input *PlayerInput) {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.inputs.PushBack(input)
}

func (q *InputQueue) Prepend(input *PlayerInput) {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.inputs.PushFront(input)
}
