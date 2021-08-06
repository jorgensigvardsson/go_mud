package main

import (
	"testing"
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
