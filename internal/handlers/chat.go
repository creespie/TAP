package handlers

import (
	"tap/internal/world"
)

func Send(p *world.Player, line string) {
	select {
	case p.Outbox <- line:
	default: // coda piena: il client non legge, la riga viene scartata
	}
}

func SendToAll(players map[string]*world.Player, message string) {
	for _, p := range players {
		Send(p, message)
	}
}

func HandleChatGlobal(sender *world.Player, message string, w *world.World) {
	message = ("EVT GLOBAL CHAT " + sender.Name + " " + message)
	Send(sender, "OK")
	SendToAll(w.Players, message)
}

func HandleChatRoom(sender *world.Player, message string, w *world.World) {
	message = ("EVT ROOM CHAT " + sender.Name + " " + message)
	Send(sender, "OK")
	SendToAll(sender.CurrentRoom.Players, message)
}

func HandleChatGroup(sender *world.Player, message string, w *world.World) {
	if sender.Group == nil {
		Send(sender, "ERR 401 NOT_IN_GROUP")
		return
	}
	message = ("EVT GROUP CHAT " + sender.Name + " " + message)
	Send(sender, "OK")
	SendToAll(sender.Group.Players, message)
}
