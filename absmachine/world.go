package absmachine

type Direction int

const (
	DIR_NORTH Direction = iota
	DIR_SOUTH
	DIR_EAST
	DIR_WEST
	DIR_UP
	DIR_DOWN
)

var OppositeDirections = []Direction{
	// DIR_NORTH = 0
	DIR_SOUTH,
	// DIR_SOUTH = 1
	DIR_NORTH,
	// DIR_EAST = 2
	DIR_WEST,
	// DIR_WEST = 3
	DIR_EAST,
	// DIR_UP = 4
	DIR_DOWN,
	// DIR_DOWN = 5
	DIR_UP,
}

const NUM_DIR = 6

type PlayerState uint32

const (
	PS_STANDING PlayerState = 1 << iota
	PS_LOGGED_IN
	PS_BUSY
)

type PlayerClass int

const (
	PC_Warrior PlayerClass = iota
	PC_Thief
	PC_Cleric
	PC_Wizard
)

type Room struct {
	Title         string
	Description   string
	Players       []*Player
	Mobs          []*Mob
	Objects       []*Object
	AdjacentRooms [NUM_DIR]*Room
	World         *World
}

type Player struct {
	Name        string
	Description string
	Room        *Room
	World       *World
	Health      int
	Mana        int
	Level       int
	State       PlayerState
	Class       PlayerClass
}

type Mob struct {
	Name            string
	Description     string
	Room            *Room
	World           *World
	RoomDescription string
}

type World struct {
	StartRoom *Room
	Rooms     []*Room
	Players   []*Player
	Mobs      []*Mob
	Objects   []*Object
}

type Object struct {
	Name            string
	Description     string
	Room            *Room
	World           *World
	RoomDescription string
}

type RelocatableToRoom interface {
	RelocateToRoom(room *Room) *LowLevelOpsError
}

type DirectionMovable interface {
	Move(direction Direction) *LowLevelOpsError
}

func (ps PlayerState) HasFlag(f PlayerState) bool { return f&ps != 0 }
func (ps *PlayerState) SetFlag(f PlayerState)     { *ps |= f }
func (ps *PlayerState) ClearFlag(f PlayerState)   { *ps &= ^f }
func (ps *PlayerState) ToggleFlag(f PlayerState)  { *ps ^= f }
