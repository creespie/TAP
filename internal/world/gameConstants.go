package world

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
	MaxInventoryItems = 20 // Design choice
	MaxGroupPlayers   = 4  // Design choice
)
