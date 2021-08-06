package main

import (
	"testing"

	"github.com/jorgensigvardsson/gomud/absmachine"
)

func Test_AppendWorks(t *testing.T) {
	queue := NewInputQueue()

	queue.Append(&PlayerInput{
		tick:  1,
		input: "one",
	})

	queue.Append(&PlayerInput{
		tick:  2,
		input: "two",
	})

	i := queue.inputs.Front()

	if i.Value.(*PlayerInput).tick != 1 || i.Value.(*PlayerInput).input != "one" {
		t.Error("First element not tick = 1 && input == \"one\"")
	}

	i = i.Next()

	if i.Value.(*PlayerInput).tick != 2 || i.Value.(*PlayerInput).input != "two" {
		t.Error("First element not tick = 2 && input == \"two\"")
	}

	i = i.Next()

	if i != nil {
		t.Error("There are more than 2 elements in the queue!")
	}
}

func Test_PrependWorks(t *testing.T) {
	queue := NewInputQueue()

	queue.Prepend(&PlayerInput{
		tick:  2,
		input: "two",
	})

	queue.Prepend(&PlayerInput{
		tick:  1,
		input: "one",
	})

	i := queue.inputs.Front()

	if i.Value.(*PlayerInput).tick != 1 || i.Value.(*PlayerInput).input != "one" {
		t.Error("First element not tick = 1 && input == \"one\"")
	}

	i = i.Next()

	if i.Value.(*PlayerInput).tick != 2 || i.Value.(*PlayerInput).input != "two" {
		t.Error("First element not tick = 2 && input == \"two\"")
	}

	i = i.Next()

	if i != nil {
		t.Error("There are more than 2 elements in the queue!")
	}
}

func Test_ForEachUntilTick_EmptyQueue_NothingHappens(t *testing.T) {
	queue := NewInputQueue()

	callCount := 0
	queue.ForEachUntilTick(1, func(playerInput *PlayerInput) {
		callCount++
	})

	if callCount > 0 {
		t.Error("Lambda called!")
	}
}

func Test_ForEachUntilTick_OlderItemsAreProcessedAndRemoved(t *testing.T) {
	player1 := &absmachine.Player{}
	player2 := &absmachine.Player{}
	queue := NewInputQueue()

	queue.Append(&PlayerInput{
		tick:   1,
		input:  "input1",
		player: player1,
	})

	queue.Append(&PlayerInput{
		tick:   2,
		input:  "input2",
		player: player1,
	})

	queue.Append(&PlayerInput{
		tick:   1,
		input:  "input1",
		player: player2,
	})

	processedInputs := make([]*PlayerInput, 0, 2)

	queue.ForEachUntilTick(1, func(playerInput *PlayerInput) {
		processedInputs = append(processedInputs, playerInput)
	})

	if len(processedInputs) != 2 {
		t.Errorf("Expected 2 player inputs to be processed, but %v were processed", len(processedInputs))
		return
	}

	if processedInputs[0].tick != 1 || processedInputs[0].player != player1 {
		t.Error("Expected player one @ tick 1 to be processed first, but got: ", processedInputs[0])
	}

	if processedInputs[1].tick != 1 || processedInputs[1].player != player2 {
		t.Error("Expected player two @ tick 1 to be processed first, but got: ", processedInputs[0])
	}

	i := queue.inputs.Front()

	if i.Value.(*PlayerInput).tick != 2 || i.Value.(*PlayerInput).player != player1 {
		t.Error("Expected player 1 @ tick 2 to be still in the queue, but got: ", i.Value)
	}

	i = i.Next()

	if i != nil {
		t.Error("There are more than 1 elements left in the queue!")
	}
}

func Test_ForEachUntilTick_CommandsForSamePlayer_SameTick_LaterArePostpnedToNextTick(t *testing.T) {
	player1 := &absmachine.Player{}
	queue := NewInputQueue()

	queue.Append(&PlayerInput{
		tick:   1,
		input:  "input1",
		player: player1,
	})

	queue.Append(&PlayerInput{
		tick:   1,
		input:  "input2",
		player: player1,
	})

	queue.Append(&PlayerInput{
		tick:   1,
		input:  "input3",
		player: player1,
	})

	processedInputs := make([]*PlayerInput, 0)

	queue.ForEachUntilTick(1, func(playerInput *PlayerInput) {
		processedInputs = append(processedInputs, playerInput)
	})

	if len(processedInputs) != 1 {
		t.Errorf("Expected 1 player inputs to be processed, but %v were processed", len(processedInputs))
		return
	}

	if processedInputs[0].tick != 1 || processedInputs[0].input != "input1" {
		t.Error("Expected input1 @ tick 1 to be processed first, but got: ", processedInputs[0])
	}

	i := queue.inputs.Front()

	if i.Value.(*PlayerInput).tick != 2 || i.Value.(*PlayerInput).input != "input2" {
		t.Error("Expected input2 @ tick 2 to be still in the queue, but got: ", i.Value)
	}

	i = i.Next()

	if i.Value.(*PlayerInput).tick != 2 || i.Value.(*PlayerInput).input != "input3" {
		t.Error("Expected input3 @ tick 2 to be still in the queue, but got: ", i.Value)
	}

	i = i.Next()

	if i != nil {
		t.Error("There are more than 2 elements left in the queue!")
	}
}
