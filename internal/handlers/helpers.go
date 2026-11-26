package handlers

import (
	"strings"
	"tap/internal/world"
)

// general function to remove every instance of a player in the world
func RemovePlayer(w *world.World, player *world.Player) {
	delete(w.Players, player.Name)
	delete(player.CurrentRoom.Players, player.Name)
	if player.Group != nil {
		delete(player.Group.Players, player.Name)
	}
}

// general function to move player to next room
func MovePlayer(player *world.Player, room *world.Room) {
	delete(player.CurrentRoom.Players, player.Name)
	room.Players[player.Name] = player
	player.CurrentRoom = room
}

// FindItem returns the first item in objects matching arg, or nil.
// arg can be the full id ("item.herbs"), the id without the "item."
// prefix ("herbs") or the display name ("Healing Herbs"); case is ignored.
func FindItem(objects map[int]*world.Object, arg string) *world.Object {
	for _, obj := range objects {
		if strings.EqualFold(obj.Id, arg) ||
			strings.EqualFold(strings.TrimPrefix(obj.Id, "item."), arg) ||
			strings.EqualFold(obj.Name, arg) {
			return obj
		}
	}
	return nil
}

func FindNpc(room *world.Room, arg string) *world.Npc {
	for _, npc := range room.Npc {
		if !npc.IsAlive {
			continue
		}
		if strings.EqualFold(npc.Id, arg) ||
			strings.EqualFold(strings.TrimPrefix(npc.Id, "npc."), arg) ||
			strings.EqualFold(npc.Name, arg) {
			return npc
		}
	}
	return nil
}
