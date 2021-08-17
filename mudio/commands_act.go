package mudio

import (
	"github.com/jorgensigvardsson/gomud/absmachine"
	"github.com/jorgensigvardsson/gomud/lang"
)

/**** Command: Look ****/
type CommandLook struct {
	args []string
}

func NewCommandLook(args []string) (Command, CommandRequirementsEvaluator) {
	return &CommandLook{args}, nil
}

func (command *CommandLook) Execute(context *CommandContext) (CommandResult, *CommandError) {
	return doLook(context)
}

func doLook(context *CommandContext, args ...string) (CommandResult, *CommandError) {
	if context.Player.Room == nil {
		return CommandResult{Output: "It would seem you're not in a room, but in a void. What happened!?"}, nil
	}

	b := buffer{}

	// Show the title of the room
	b.Println(context.Player.Room.Title)

	// Show the description of the room (and indent first line)
	b.Printf("   ")
	b.Println(context.Player.Room.Description)

	for _, player := range context.Player.Room.Players {
		if player != context.Player {
			// TODO: Show whether the other player is standing, sleeping, or whatever
			b.Printlnf("%v is standing here", player.Name)
		}
	}

	for _, mob := range context.Player.World.Mobs {
		if mob.RoomDescription != "" {
			b.Printlnf(mob.RoomDescription)
		} else {
			b.Printlnf("%v %v is here.", lang.IndefiniteArticleFor(mob.Name), mob.Name)
		}
	}

	for _, object := range context.Player.World.Objects {
		if object.RoomDescription != "" {
			b.Printlnf(object.RoomDescription)
		} else {
			b.Printlnf("%v %v is lying on the ground.", lang.IndefiniteArticleFor(object.Name), object.Name)
		}
	}

	b.Println("Obvious exits:")
	hasAdjacentRoom := false
	for d, r := range context.Player.Room.AdjacentRooms {
		if r != nil {
			b.Printlnf("%-10s - %s", lang.DirectionName(absmachine.Direction(d)), r.Title)
			hasAdjacentRoom = true
		}
	}

	if !hasAdjacentRoom {
		b.Println("NONE - YOU ARE TRAPPED!")
	}

	return CommandResult{Output: b.ToString()}, nil
}
