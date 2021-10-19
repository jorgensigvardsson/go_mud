package mudio

import (
	"fmt"
	"strings"

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
	if len(command.args) == 0 {
		return lookRoom(context)
	}

	mob, obj, player := findTargets(context.Player.Room, command.args[0])

	nilCount := 0
	if mob == nil {
		nilCount++
	}

	if obj == nil {
		nilCount++
	}

	if player == nil {
		nilCount++
	}

	if nilCount == 3 {
		return CommandResult{}, &CommandError{fmt.Sprintf("Can't find %v in the room...", command.args[0])}
	} else if nilCount != 2 {
		// TODO: How to let user disambiguate?
		return CommandResult{}, &CommandError{fmt.Sprintf("There are more than one thing in the room called %v...", command.args[0])}
	}

	if mob != nil {
		return CommandResult{Output: mob.Description}, nil
	}

	if obj != nil {
		return CommandResult{Output: obj.Description}, nil
	}

	return CommandResult{Output: player.Description}, nil
}

func findTargets(room *absmachine.Room, target string) (*absmachine.Mob, *absmachine.Object, *absmachine.Player) {
	targetLowerCase := strings.ToLower(target)

	var mob *absmachine.Mob = nil
	var obj *absmachine.Object = nil
	var player *absmachine.Player = nil

	for _, aMob := range room.Mobs {
		if strings.HasPrefix(strings.ToLower(aMob.Name), targetLowerCase) {
			mob = aMob
			break
		}
	}

	for _, anObj := range room.Objects {
		if strings.HasPrefix(strings.ToLower(anObj.Name), targetLowerCase) {
			obj = anObj
			break
		}
	}

	for _, aPlayer := range room.Players {
		if strings.HasPrefix(strings.ToLower(aPlayer.Name), targetLowerCase) {
			player = aPlayer
			break
		}
	}

	return mob, obj, player
}

func lookRoom(context *CommandContext, args ...string) (CommandResult, *CommandError) {
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
