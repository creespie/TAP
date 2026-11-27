package handlers

import (
	"fmt"
	"strconv"

	"tap/internal/world"
)

//deals damage to npc, returns true if the npc dies
func playerHits(w *world.World, npc *world.Npc, pStats world.Stats, player *world.Player) bool {
	crit := w.Rand.Intn(100) < world.CriticalChance+pStats.Luck
	if crit {
		Send(player, "Critical Attack! "+strconv.Itoa(pStats.Attack*world.CriticalMultiplier)+"DMG!")
		npc.Hp -= pStats.Attack * world.CriticalMultiplier
	} else {
		Send(player, "You attacked for "+strconv.Itoa(pStats.Attack)+"DMG!")
		npc.Hp -= pStats.Attack
	}
	if npc.Hp > 0 {
		return false
	}
	npc.IsAlive = false
	return true
}

//deals damage to player, returns true if the player dies
func npcHits(w *world.World, npc *world.Npc, pStats world.Stats, player *world.Player) bool {
	crit := w.Rand.Intn(100) < world.CriticalChance
	action := npc.Action[w.Rand.Intn(len(npc.Action))]
	if crit {
		Send(player, action.Sentence)
		Send(player, "The "+npc.Name+" attack was critical!")
		player.Stats.Health -= action.Damage * world.CriticalMultiplier
	} else {
		Send(player, action.Sentence)
		player.Stats.Health -= action.Damage
	}
	if player.Stats.Health > 0 {
		return false
	}
	return true
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

func respawnPlayer(player *world.Player, w *world.World) {
	//chance of droping the weapon
	lost := w.Rand.Float64() < world.WeaponDropChance
	weapon := player.Equipment[world.Weapon]
	if lost && weapon != nil {
		player.CurrentRoom.Objects[weapon.ItemNumber] = weapon
		delete(player.Equipment, world.Weapon)
	}
	player.Stats.Health = world.PlayerRespawnHealth
	player.CurrentRoom = w.RespawnRoom
	player.Target = nil
}

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
		Send(player, "ERR 405 ALREADY_IN_COMBAT")
		return
	}
	player.Target = npc
	pStats := PlayerStats(player)
	//checks who attacks first
	if float64(pStats.Speed)*w.Rand.Float64() > float64(npc.Speed)*w.Rand.Float64() {
		killed := playerHits(w, npc, pStats, player)
		if killed {
			Send(player, "OK killed "+npc.Name)
			registerKill(player, npc.Id)
			player.Target = nil
			return
		}
		pKilled := npcHits(w, npc, pStats, player)
		if pKilled {
			Send(player, "ARGH! You died!")
			respawnPlayer(player, w)
			return
		}
	} else {
		pKilled := npcHits(w, npc, pStats, player)
		if pKilled {
			Send(player, "ARGH! You died!")
			respawnPlayer(player, w)
			return
		}
		killed := playerHits(w, npc, pStats, player)
		if killed {
			Send(player, "OK killed "+npc.Name)
			registerKill(player, npc.Id)
			player.Target = nil
			return
		}
	}
}
