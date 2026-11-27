package handlers

import (
	"encoding/json"
	"fmt"
	"os"

	"tap/internal/world"
)

// attackResponse is the reply to ATTACK, in the format of RFC 42TAP 5.4.5.
type attackResponse struct {
	AttackerHP int    `json:"attacker_hp"`
	TargetHP   int    `json:"target_hp"`
	Damage     int    `json:"damage"`
	Status     string `json:"status"` // "combat", "victory" or "defeated"
}

// combatEvent tells everyone in room what happened in a fight.
func combatEvent(room *world.Room, text string) {
	SendToAll(room.Players, "EVT ROOM COMBAT "+text)
}

// playerHits deals the player's damage to npc and returns it.
func playerHits(w *world.World, player *world.Player, npc *world.Npc, s world.Stats) int {
	dmg := s.Attack
	crit := w.Rand.Intn(100) < world.CriticalChance+s.Luck
	if crit {
		dmg *= world.CriticalMultiplier
	}
	if dmg < 1 {
		dmg = 1
	}
	npc.Hp -= dmg

	text := fmt.Sprintf("%s hits %s for %d", player.Name, npc.Name, dmg)
	if crit {
		text += " (critical)"
	}
	combatEvent(player.CurrentRoom, text)
	return dmg
}

// npcHits deals the damage of a random npc action to the player and returns it.
// The player's defense is subtracted; every hit deals at least 1.
func npcHits(w *world.World, player *world.Player, npc *world.Npc, s world.Stats) int {
	if len(npc.Action) == 0 {
		return 0
	}
	action := npc.Action[w.Rand.Intn(len(npc.Action))]
	dmg := action.Damage
	crit := w.Rand.Intn(100) < world.CriticalChance
	if crit {
		dmg *= world.CriticalMultiplier
	}
	dmg -= s.Defense
	if dmg < 1 {
		dmg = 1
	}
	player.Stats.Health -= dmg

	text := fmt.Sprintf("%s %s takes %d damage", action.Sentence, player.Name, dmg)
	if crit {
		text += " (critical)"
	}
	combatEvent(player.CurrentRoom, text)
	return dmg
}

// registerKill advances every active kill quest of player that targets npcID.
func registerKill(player *world.Player, npcID string) {
	for id, q := range player.Quests {
		if q.Type == world.Kill && !q.Completed && q.Target == npcID {
			q.Progress++
			player.Quests[id] = q
			Send(player, fmt.Sprintf("EVT QUEST PROGRESS %s %d/%d", q.Id, q.Progress, q.Count))
		}
	}
}

// creditKill gives the kill to the killer, or to every member of the
// killer's group (the killer included) if the killer is in a group.
func creditKill(killer *world.Player, npcID string) {
	if killer.Group == nil {
		registerKill(killer, npcID)
		return
	}
	for _, member := range killer.Group.Players {
		registerKill(member, npcID)
	}
}

// npcDies marks npc as dead, ends the fight for everyone who was fighting it
// and credits the kill to the killer's group.
func npcDies(killer *world.Player, npc *world.Npc) {
	room := killer.CurrentRoom
	npc.Hp = 0
	npc.IsAlive = false
	for _, p := range room.Players {
		if p.Target == npc {
			p.Target = nil
		}
	}
	combatEvent(room, fmt.Sprintf("%s has been defeated by %s", npc.Name, killer.Name))
	SendToAll(room.Players, "EVT ROOM NPC DEATH "+npc.Id)
	creditKill(killer, npc.Id)
}

// respawnPlayer handles a player at 0 HP: the equipped weapon may drop in the
// room, then the player reappears in the respawn room with reduced HP.
func respawnPlayer(w *world.World, player *world.Player) {
	room := player.CurrentRoom
	combatEvent(room, player.Name+" has been defeated")

	weapon := player.Equipment[world.Weapon]
	if weapon != nil && w.Rand.Float64() < world.WeaponDropChance {
		delete(player.Equipment, world.Weapon)
		room.Objects[weapon.ItemNumber] = weapon
		SendToAll(room.Players, "EVT ROOM ITEM DROP "+player.Name+" "+weapon.Id)
	}

	player.Target = nil
	player.Stats.Health = world.PlayerRespawnHealth
	MovePlayer(player, w.RespawnRoom)
}

// HandleAttack plays one turn of combat against an npc in the player's room.
// The fastest fighter (speed x random roll, rolled every turn) hits first;
// the other hits back if still alive. The reply is always one OK line.
func HandleAttack(w *world.World, player *world.Player, npcName string) {
	npc := FindNpc(player.CurrentRoom, npcName)
	if npc == nil {
		Send(player, "ERR 404 NPC_NOT_FOUND")
		return
	}
	if npc.Behaviour != world.Aggressive {
		Send(player, "ERR 405 NPC_NOT_HOSTILE")
		return
	}
	if player.Target != nil && player.Target != npc {
		Send(player, "ERR 407 IN_COMBAT")
		return
	}
	player.Target = npc
	s := PlayerStats(player)

	dealt := 0
	playerFirst := float64(s.Speed)*w.Rand.Float64() > float64(npc.Speed)*w.Rand.Float64()
	if playerFirst {
		dealt = playerHits(w, player, npc, s)
		if npc.Hp > 0 {
			npcHits(w, player, npc, s)
		}
	} else {
		npcHits(w, player, npc, s)
		if player.Stats.Health > 0 {
			dealt = playerHits(w, player, npc, s)
		}
	}

	status := "combat"
	if npc.Hp <= 0 {
		status = "victory"
		npcDies(player, npc)
	} else if player.Stats.Health <= 0 {
		status = "defeated"
		respawnPlayer(w, player)
	}

	targetHP := npc.Hp
	if targetHP < 0 {
		targetHP = 0
	}
	data, err := json.Marshal(attackResponse{
		AttackerHP: player.Stats.Health, // after a defeat: the HP the player respawned with
		TargetHP:   targetHP,
		Damage:     dealt,
		Status:     status,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "attack: marshal:", err)
		Send(player, "ERR 500 INTERNAL_ERROR")
		return
	}
	Send(player, "OK "+string(data))
}