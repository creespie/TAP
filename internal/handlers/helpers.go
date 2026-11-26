package handlers

import(
	"tap/internal/world"
)

//general function to remove every instance of a player in the world
func RemovePlayer(w *world.World, player *world.Player) {
	delete(w.Players, player.Name)
	delete(player.CurrentRoom.Players, player.Name)
	if player.Group != nil {
		delete(player.Group.Players, player.Name)
	}
}

//general function to move player to next room
func MovePlayer(player *world.Player, room *world.Room){
	delete(player.CurrentRoom.Players, player.Name)
	room.Players[player.Name] = player
	player.CurrentRoom = room
}
