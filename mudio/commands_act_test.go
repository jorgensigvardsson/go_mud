package mudio

import (
	"strings"
	"testing"

	"github.com/jorgensigvardsson/gomud/absmachine"
)

func Test_look_at_empty_room_no_exits(t *testing.T) {
	command := CommandLook{args: []string{}}
	room := absmachine.Room{Title: "The room", Description: "A description"}
	world := absmachine.World{}
	player := absmachine.Player{Room: &room, World: &world}
	context := CommandContext{Player: &player}

	result, err := command.Execute(&context)

	if err != nil {
		t.Errorf("An unexpected error occurred: %v", *err)
	}

	if !strings.Contains(result.Output, "The room") {
		t.Error("Title is missing in the room description!")
	}

	if !strings.Contains(result.Output, "NONE - YOU ARE TRAPPED") {
		t.Error("There should be no exits!")
	}
}

func Test_look_at_empty_room_with_exits(t *testing.T) {
	command := CommandLook{args: []string{}}
	room := absmachine.Room{Title: "The room", Description: "A description"}
	nextRoom := absmachine.Room{Title: "The next room"}
	world := absmachine.World{Rooms: []*absmachine.Room{&room, &nextRoom}}
	room.Connect(&nextRoom, absmachine.DIR_NORTH)
	player := absmachine.Player{Room: &room, World: &world}
	context := CommandContext{Player: &player}

	result, err := command.Execute(&context)

	if err != nil {
		t.Errorf("An unexpected error occurred: %v", *err)
	}

	if !strings.Contains(result.Output, "North      - The next room") {
		t.Errorf("The next room is not due north!, output = %v", result.Output)
	}
}

func Test_look_at_room_with_mobs(t *testing.T) {
	command := CommandLook{args: []string{}}
	room := absmachine.Room{Title: "The room", Description: "A description"}
	nextRoom := absmachine.Room{Title: "The next room"}
	world := absmachine.World{Rooms: []*absmachine.Room{&room, &nextRoom}}
	mob := absmachine.Mob{RoomDescription: "The mob is here."}
	room.Mobs = append(room.Mobs, &mob)
	room.Connect(&nextRoom, absmachine.DIR_NORTH)
	player := absmachine.Player{Room: &room, World: &world}
	context := CommandContext{Player: &player}

	result, err := command.Execute(&context)

	if err != nil {
		t.Errorf("An unexpected error occurred: %v", *err)
	}

	if !strings.Contains(result.Output, "The mob is here.") {
		t.Errorf("The mob's room description cannot be found in output!, output = %v", result.Output)
	}
}

func Test_look_at_non_existing_thing(t *testing.T) {
	command := CommandLook{args: []string{"thing"}}
	room := absmachine.Room{Title: "The room", Description: "A description"}
	world := absmachine.World{Rooms: []*absmachine.Room{&room}}
	player := absmachine.Player{Room: &room, World: &world}
	context := CommandContext{Player: &player}

	result, err := command.Execute(&context)

	if err == nil {
		t.Error("The expected error did not occurr")
	} else if err.Error() != "Can't find thing in the room..." {
		t.Errorf("Unexpected error occurred: %v", err)
	}

	if result.Output != "" {
		t.Errorf("Unexpected output: %v", result.Output)
	}
}

func Test_look_at_mob(t *testing.T) {
	command := CommandLook{args: []string{"thing"}}
	room := absmachine.Room{Title: "The room", Description: "A description"}
	world := absmachine.World{Rooms: []*absmachine.Room{&room}}
	mob := absmachine.Mob{Name: "thing", Description: "The thing is here."}
	room.Mobs = append(room.Mobs, &mob)
	player := absmachine.Player{Room: &room, World: &world}
	context := CommandContext{Player: &player}

	result, err := command.Execute(&context)

	if err != nil {
		t.Errorf("Unexpected error occurred: %v", err)
	}

	if result.Output != "The thing is here." {
		t.Errorf("Unexpected output: %v", result.Output)
	}
}
