package types

func (world World) MovePlayerToRoom(player *Player, room *Room) {
	if room == player.Room {
		return
	}

	if player.Room != nil {
		removePlayerFromRoom(player.Room, player)
	} else {
		room.Players = append(room.Players, player)
		player.Room = room
	}
}

func indexOfRoomPlayer(room *Room, player *Player) int {
	for index, v := range room.Players {
		if player == v {
			return index
		}
	}

	return -1
}

func removePlayerFromRoom(room *Room, player *Player) {
	index := indexOfRoomPlayer(room, player)
	if index < 0 {
		panic("Player was not in the room!")
	}

	room.Players = append(room.Players[:index], room.Players[index+1:]...)
	player.Room = nil
}
