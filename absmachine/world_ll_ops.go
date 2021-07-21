package absmachine

// Adds a set of rooms to a world. None of the rooms may be associated with a world already!
func (world *World) AddRooms(rooms []*Room) *LowLevelOpsError {
	// Check for inconsistencies

	// Are any of the supplied rooms already attached to a world?
	for _, room := range rooms {
		if room.World != nil {
			return &LowLevelOpsError{errorCode: ErrorIconsistency, message: "At least one room is already attached to another world!"}
		}
	}

	// Nope, let's go right ahead and add them!
	world.Rooms = append(world.Rooms, rooms...)
	for _, room := range rooms {
		room.World = world
	}
	return nil
}

// Adds a set of players to a world. None of the players may be associated with a world already!
func (world *World) AddPlayers(players []*Player) *LowLevelOpsError {
	// Check for inconsistencies

	// Are any of the supplied players already attached to a world?
	for _, player := range players {
		if player.World != nil {
			return &LowLevelOpsError{errorCode: ErrorIconsistency, message: "At least one player is already attached to another world!"}
		}
	}

	// Nope, let's go right ahead and add them!
	world.Players = append(world.Players, players...)
	for _, player := range players {
		player.World = world
	}

	return nil
}

// Moves a player to a room
func (world World) MovePlayerToRoom(player *Player, room *Room) *LowLevelOpsError {
	if room == player.Room {
		return nil
	}

	if player.Room != nil {
		err := removePlayerFromRoom(player.Room, player)
		if err != nil {
			return err
		}
	}

	room.Players = append(room.Players, player)
	player.Room = room
	return nil
}

// Moves the player in a specific direction
func (player *Player) MovePlayer(direction Direction) *LowLevelOpsError {
	// Sanity checks!
	if player.Room == nil {
		return &LowLevelOpsError{errorCode: ErrorIconsistency, message: "Player has no reference to a room!"}
	}

	if player.Room.World == nil {
		return &LowLevelOpsError{errorCode: ErrorIconsistency, message: "Player has no reference to World via its referenced room!"}
	}

	// Does a room exist in the direction of the player?
	if player.Room.AdjacentRooms[direction] == nil {
		return &LowLevelOpsError{errorCode: ErrorInvalidDirection, message: "Player cannot move in specified direction!"}
	}

	player.Room.World.MovePlayerToRoom(player, player.Room.AdjacentRooms[direction])
	return nil
}

func (room *Room) Connect(otherRoom *Room, direction Direction) *LowLevelOpsError {
	if room.AdjacentRooms[direction] != nil {
		return &LowLevelOpsError{errorCode: ErrorIconsistency, message: "Room is already connected to another room in specified direction"}
	}

	room.AdjacentRooms[direction] = otherRoom
	return nil
}

func indexOfRoomPlayer(room *Room, player *Player) int {
	for index, v := range room.Players {
		if player == v {
			return index
		}
	}

	return -1
}

func removePlayerFromRoom(room *Room, player *Player) *LowLevelOpsError {
	index := indexOfRoomPlayer(room, player)
	if index < 0 {
		return &LowLevelOpsError{errorCode: ErrorIconsistency, message: "Player was not in room's list of players!"}
	}

	room.Players = append(room.Players[:index], room.Players[index+1:]...)
	player.Room = nil
	return nil
}
