package handlers

import (
	"fmt"

	"tap/internal/world"
)

func SendToAll(players map[string]*world.Player, message string) {
	for _, p := range players {
		fmt.Fprintln(p.Connection, message)
	}
	fmt.Fprintln(players[name].Connection, "OK")
}

func handleChatGlobal(sender *world.Player, message string, w *world.World) {
	message = ("EVT GLOBAL CHAT " + sender.Name + " " + message)
	SendToAll(w.Players, message)
}

func handleChatRoom(sender *world.Player, message string, w *world.World) {
	message = ("EVT ROOM CHAT " + sender.Name + " " + message)
	SendToAll(sender.CurrentRoom.Players, message)
}

func handleChatGroup(sender *world.Player, message string, w *world.World) {
	message = ("EVT GROUP CHAT " + sender.Name + " " + message)
	SendToAll(sender.Group.Players, message)
}
