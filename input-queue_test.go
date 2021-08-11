package main

import (
	"fmt"
	"testing"

	"github.com/jorgensigvardsson/gomud/absmachine"
)

func Test_showNormalPrompt(t *testing.T) {
	conn := &FakeTelnetConnection{}
	player := absmachine.Player{
		Health: 103,
		Mana:   43,
	}

	showNormalPrompt(conn, &player)

	if conn.text != "[H:103] [M:43] > " {
		t.Errorf("Unexpected prompt: %v", conn.text)
	}
}

type FakeTelnetConnection struct {
	text string
}

func (conn *FakeTelnetConnection) ReadLine() (line string, err error) {
	panic("ReadLine not implemented")
}

func (conn *FakeTelnetConnection) WriteLine(line string) error {
	conn.text += line + "\r\n"
	return nil
}

func (conn *FakeTelnetConnection) WriteLinef(line string, args ...interface{}) error {
	conn.text += fmt.Sprintf(line, args...) + "\r\n"
	return nil
}

func (conn *FakeTelnetConnection) WriteString(text string) error {
	conn.text += text
	return nil
}

func (conn *FakeTelnetConnection) WriteStringf(text string, args ...interface{}) error {
	conn.text += fmt.Sprintf(text, args...)
	return nil
}

func (conn *FakeTelnetConnection) EchoOff() error { panic("EchoOff not implemented") }
func (conn *FakeTelnetConnection) EchoOn() error  { panic("EchoOn not implemented") }
func (conn *FakeTelnetConnection) Close() error   { panic("Close not implemented") }

func Test_Append_PlayerLimitIsRespected(t *testing.T) {
	q := NewInputQueue(1, 1)
	p1 := absmachine.NewPlayer()
	p2 := absmachine.NewPlayer()
	errorChannel1 := make(chan error, 10)
	errorChannel2 := make(chan error, 10)

	p1.Name = "p1"
	p2.Name = "p2"

	q.Append(
		&PlayerInput{
			player:             p1,
			text:               "cmd",
			errorReturnChannel: errorChannel1,
		},
	)

	q.Append(
		&PlayerInput{
			player:             p2,
			text:               "cmd",
			errorReturnChannel: errorChannel2,
		},
	)

	select {
	case err := <-errorChannel1:
		t.Errorf("Unexpected error on channel 1: %v", err)
	default:
	}

	select {
	case err := <-errorChannel2:
		if err != ErrTooManyPlayers {
			t.Errorf("Unexpected error: %v", err)
		}
	default:
		t.Errorf("Unexpectedly, there was no error on channel 2!")
	}
}

func Test_Append_PlayerInputLimitIsRespected(t *testing.T) {
	q := NewInputQueue(1, 1)
	p := absmachine.NewPlayer()
	errorChannel := make(chan error, 10)

	p.Name = "p1"

	q.Append(
		&PlayerInput{
			player:             p,
			text:               "cmd 1",
			errorReturnChannel: errorChannel,
		},
	)

	q.Append(
		&PlayerInput{
			player:             p,
			text:               "cmd 2",
			errorReturnChannel: errorChannel,
		},
	)

	select {
	case err := <-errorChannel:
		if err != ErrTooMuchInput {
			t.Errorf("Unexpected error: %v", err)
		}
	default:
		t.Errorf("Unexpectedly, there was no error on channel!")
	}
}
