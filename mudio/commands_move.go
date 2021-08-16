package mudio

import (
	"github.com/jorgensigvardsson/gomud/absmachine"
	"github.com/jorgensigvardsson/gomud/lang"
)

/**** Command: Move ****/
type CommandMove struct {
	direction absmachine.Direction
}

func NewCommandMoveNorth(args []string) Command {
	return &CommandMove{absmachine.DIR_NORTH}
}

func NewCommandMoveSouth(args []string) Command {
	return &CommandMove{absmachine.DIR_SOUTH}
}

func NewCommandMoveEast(args []string) Command {
	return &CommandMove{absmachine.DIR_EAST}
}

func NewCommandMoveWest(args []string) Command {
	return &CommandMove{absmachine.DIR_WEST}
}

func NewCommandMoveUp(args []string) Command {
	return &CommandMove{absmachine.DIR_UP}
}

func NewCommandMoveDown(args []string) Command {
	return &CommandMove{absmachine.DIR_DOWN}
}

func (command *CommandMove) Execute(context *CommandContext) (CommandResult, *CommandError) {
	if context.Player.Room == nil {
		return CommandResult{}, &CommandError{"It would seem you're not in a room, but in a void. What happened!?"}
	}

	adjacentRoom := context.Player.Room.AdjacentRooms[command.direction]
	if adjacentRoom == nil {
		return CommandResult{}, &CommandError{"You can't go that way."}
	}

	err := context.Player.Move(command.direction)
	if err != nil {
		context.Logger.Printlnf("Can't go %v, error: %v", lang.DirectionName(command.direction), err)
		return CommandResult{}, &CommandError{"You can't go that way."}
	}

	return CommandResult{}, nil
}
