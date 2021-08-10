package main

import (
	"testing"

	"github.com/jorgensigvardsson/gomud/absmachine"
)

func Test_AppendWorks(t *testing.T) {
	queue := NewInputQueue()

	queue.Append(
		&PlayerInputOrCommand{
			input: "one",
		},
	)

	queue.Tick()

	queue.Append(
		&PlayerInputOrCommand{
			input: "two",
		},
	)

	i := queue.inputs.Front()

	if i.Value.(*QueueEntry).tick != 0 || i.Value.(*QueueEntry).inputOrCommand.input != "one" {
		t.Error("First element not tick = 0 && input == \"one\"")
	}

	i = i.Next()

	if i.Value.(*QueueEntry).tick != 1 || i.Value.(*QueueEntry).inputOrCommand.input != "two" {
		t.Error("First element not tick = 1 && input == \"two\"")
	}

	i = i.Next()

	if i != nil {
		t.Error("There are more than 2 elements in the queue!")
	}
}

func Test_ForEachCurrentTick_EmptyQueue_NothingHappens(t *testing.T) {
	queue := NewInputQueue()

	callCount := 0
	queue.ForEachCurrentTick(func(playerInput *PlayerInputOrCommand) {
		callCount++
	})

	if callCount > 0 {
		t.Error("Lambda called!")
	}
}

func Test_ForEachCurrentTick_OlderItemsAreProcessedAndRemoved(t *testing.T) {
	player1 := &absmachine.Player{}
	player2 := &absmachine.Player{}
	queue := NewInputQueue()

	queue.Append(
		&PlayerInputOrCommand{
			input:  "input1",
			player: player1,
		},
	)

	queue.Append(
		&PlayerInputOrCommand{
			input:  "input1",
			player: player2,
		},
	)

	queue.Tick()

	queue.Append(
		&PlayerInputOrCommand{
			input:  "input2",
			player: player1,
		},
	)

	processedInputs := make([]*PlayerInputOrCommand, 0, 2)

	queue.ForEachCurrentTick(func(playerInput *PlayerInputOrCommand) {
		processedInputs = append(processedInputs, playerInput)
	})

	if len(processedInputs) != 2 {
		t.Errorf("Expected 2 player inputs to be processed, but %v were processed", len(processedInputs))
		return
	}

	if processedInputs[0].player != player1 {
		t.Error("Expected player one @ tick 1 to be processed first, but got: ", processedInputs[0])
	}

	if processedInputs[1].player != player2 {
		t.Error("Expected player two @ tick 1 to be processed first, but got: ", processedInputs[0])
	}

	i := queue.inputs.Front()

	if i.Value.(*QueueEntry).tick != 2 || i.Value.(*QueueEntry).inputOrCommand.player != player1 {
		t.Error("Expected player 1 @ tick 2 to be still in the queue, but got: ", i.Value)
	}

	i = i.Next()

	if i != nil {
		t.Error("There are more than 1 elements left in the queue!")
	}
}

func Test_ForEachCurrentTick_CommandsForSamePlayer_SameTick_LaterArePostponedToNextTick(t *testing.T) {
	player1 := &absmachine.Player{}
	queue := NewInputQueue()

	queue.Append(
		&PlayerInputOrCommand{
			input:  "input1",
			player: player1,
		},
	)

	queue.Append(
		&PlayerInputOrCommand{
			input:  "input2",
			player: player1,
		},
	)

	queue.Append(
		&PlayerInputOrCommand{
			input:  "input3",
			player: player1,
		},
	)

	processedInputs := make([]*PlayerInputOrCommand, 0)

	queue.ForEachCurrentTick(func(playerInput *PlayerInputOrCommand) {
		processedInputs = append(processedInputs, playerInput)
	})

	if len(processedInputs) != 1 {
		t.Errorf("Expected 1 player inputs to be processed, but %v were processed", len(processedInputs))
		return
	}

	if processedInputs[0].input != "input1" {
		t.Error("Expected input1 @ tick 1 to be processed first, but got: ", processedInputs[0])
	}

	i := queue.inputs.Front()

	if i.Value.(*QueueEntry).tick != 1 || i.Value.(*QueueEntry).inputOrCommand.input != "input2" {
		t.Error("Expected input2 @ tick 1 to be still in the queue, but got: ", i.Value)
	}

	i = i.Next()

	if i.Value.(*QueueEntry).tick != 1 || i.Value.(*QueueEntry).inputOrCommand.input != "input3" {
		t.Error("Expected input3 @ tick 1 to be still in the queue, but got: ", i.Value)
	}

	i = i.Next()

	if i != nil {
		t.Error("There are more than 2 elements left in the queue!")
	}
}
