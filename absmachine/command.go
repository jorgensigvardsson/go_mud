package absmachine

type CommandQueue interface {
	Enqueue(command *Command)
	Dequeue() *Command
}

type Command interface {
	Execute() *LowLevelOpsError
}

type LookCommand struct {
	player       *Player
	nameOfObject string
}

func NewLookCommand(player *Player, nameOfObject string) *LookCommand {
	return &LookCommand{player, nameOfObject}
}

func (command *LookCommand) Execute() *LowLevelOpsError {
	return &LowLevelOpsError{} // TODO: Implement me!
}
