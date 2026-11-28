package handlers

import (
	"tap/internal/world"
	"time"
)

func spawnChance(w *world.World, typeID string) float64 {
	cat := w.Items[typeID].Category
	base := world.SpawnChanceOther
	switch cat {
	case world.Consumable:
		base = world.SpawnChanceConsumable
	case world.Weapon:
		base = world.SpawnChanceWeapon
	}
	p := base / (1 + float64(w.Created[typeID])/world.SpawnDecayK)
	if p < world.SpawnMinChance {
		return world.SpawnMinChance
	}
	return p
}

func trySpawn(w *world.World, room *world.Room) {
	if time.Since(room.LastSpawnCheck) < world.SpawnCooldown {
		return
	}
	room.LastSpawnCheck = time.Now()
	for _, typeID := range room.SpawnTypes {
		if countItems(room.Objects, typeID) >= room.MaxItems {
			continue
		} else if w.Rand.Float64() < spawnChance(w, typeID) {
			inst := w.NewItem(typeID)
			room.Objects[inst.ItemNumber] = inst
			SendToAll(room.Players, "EVT ROOM ITEM SPAWN "+inst.Id)
		}
	}
	for _, npc := range room.Npc {
		if npc.IsAlive || time.Since(npc.DiedAt) < world.NpcRespawnCooldown {
			continue
		} else if w.Rand.Float64() < world.NpcRespawnChance {
			npc.IsAlive = true
			npc.Hp = npc.MaxHp
			SendToAll(room.Players, "EVT ROOM NPC SPAWN "+npc.Id)
		}
	}
}
