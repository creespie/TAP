package world

import (
	"time"
)

// ==================================
// PLAYER BASE STATS
// ==================================

const (
	PlayerStartingHealth  = 100 // Required by project
	PlayerStartingSpeed   = 10  // Design choice
	PlayerStartingDefense = 0   // Design choice
	PlayerStartingLuck    = 0   // Design choice

	PlayerRespawnHealth = 50 // Design choice: reduced HP after respawn
)

// ==================================
// COMBAT
// ==================================

const (
	BaseAttackDamage   = 10 // Design choice
	CriticalChance     = 10 // Percentage
	CriticalMultiplier = 2
)

// ==================================
// GAME RULES
// ==================================

const (
	MaxInventoryItems     = 20 // Design choice
	MaxGroupPlayers       = 4  // Design choice
	WeaponDropChance      = 0.5
	DefaultRoomMaxItems   = 5
	SpawnCooldown         = 10 * time.Second
	SpawnDecayK           = 20
	SpawnMinChance        = 0.02
	NpcRespawnCooldown    = 20 * time.Second
	NpcRespawnChance      = 0.3
	SpawnChanceConsumable = 0.6  // erbe, pane: ricompaiono spesso
	SpawnChanceWeapon     = 0.04 // armi: rare
	SpawnChanceOther      = 0.2
)

// ==================================
// GAME RULES
// ==================================

const (
	OutboxSize = 64 //Design choice
)
