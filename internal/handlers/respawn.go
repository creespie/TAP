package handlers

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
	
}
