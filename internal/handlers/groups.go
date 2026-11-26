package handlers

import (
	"tap/internal/world"
	"strconv"
)

func HandleGroupCreate(w *world.World, player *world.Player){
	if player.Group != nil{
		Send(player, "ERR 402 ALREADY_IN_GROUP")
		return
	}
	group := &world.Group{
		Id: w.ItemCount,
		Players: make(map[string]*world.Player),
	}
	group.Players[player.Name] = player
	player.Group = group
	w.ItemCount++
	Send(player, "OK group=" + strconv.Itoa(player.Group.Id))
}

func HandleGroupInvite(w *world.World, player string, sender *world.Player){
	if sender.Group == nil{
		Send(sender, "404 NOT_IN_A_GROUP")
		return
	}
	invited := w.Players[player]
	if invited == nil{
		Send(sender, "404 PLAYER_NOT_FOUND")
		return
	}
	invited.Invites[strconv.Itoa(sender.Group.Id)] = sender.Group
	Send(sender, "OK")
	Send(invited, "You are now invited to group " + strconv.Itoa(sender.Group.Id))
}

func HandleGroupJoin(w *world.World, groupId string, player *world.Player) {
	group := player.Invites[groupId]
	if group == nil{
		Send(player,"ERR 404 GROUP_INVITE_NOT_FOUND")
		return
	}
	if len(group.Players) == 0{
		Send(player, "ERR 404 GROUP_WAS_DISBANDED")
		delete(player.Invites, groupId)
		return
	}
	if player.Group != nil{
		Send(player, "ERR 402 ALREADY_IN_GROUP")
		return
	}
	player.Group = group
	Send(player, "OK group=" + strconv.Itoa(group.Id))
	delete(player.Invites, groupId)
}

func HandleGroupLeave(w *world.World, player *world.Player){
	if player.Group == nil{
		Send(player, "ERR 404 NOT_IN_A_GROUP")
		return
	}
	delete(player.Group.Players, player.Name)
	player.Group = nil
	Send(player, "OK")
}
