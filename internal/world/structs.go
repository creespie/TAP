package world

import (
	"net"
	"sync"
)

type World struct {
	Mu          sync.Mutex
	Rooms       map[string]*Room
	Players     map[string]*Player
	Items       map[string]*Object // all item definitions by id (also used as reward templates)
	Npcs        map[string]*Npc    // all npcs by id
	Quests      map[string]*Quest  // all quests by id
	StartRoom   *Room              // where new players appear
	RespawnRoom *Room              // safe room for players at 0 HP
	ItemCount   int
}

func (w *World) NewItem(typeID string) *Object {
	original := w.Items[typeID]
	inst := *original // copia di tutta la struct
	inst.ItemNumber = w.ItemCount
	w.ItemCount++
	return &inst

}

type Room struct {
	Id              string
	Name            string
	Description     string
	Npc             map[string]*Npc
	Objects         map[int]*Object
	Width, height   int
	Players         map[string]*Player
	AccessibleRooms map[string]*Room // key = direction (north, south, ...)
	NpcPosition     [2]int
	ObjPosition     [2]int
}

type Npc struct {
	Id          string
	Name        string
	Description string
	Hp          int
	IsAlive     bool
	Behaviour   Behaviour
	Dialogs     map[int]string
	Action      map[int]*Action
	Quests      map[int]*Quest
}

type Quest struct {
	Id           string
	Name         string
	Description  string
	Step         int
	CompleteStep int
	Instruction  map[int]string
	Reward       Object // create a new copy and place it in the player inventory
	Progression  map[int]*Player
	Unique       bool
}

type Object struct {
	Id          string
	ItemNumber  int
	Name        string
	Obtainable  bool
	Category    Category
	Stat        int
	Description string
}

type Player struct {
	Connection  net.Conn
	Name        string
	Stats       Stats
	Inventory   map[int]*Object
	Equipment   Equipment
	Quests      map[string]Quest
	Group       *Group
	Invites     map[string]*Group
	CurrentRoom *Room
	Outbox      chan string
	TalkIndex   map[string]int
}

type Equipment struct {
	Weapon     *Object
	Helmet     *Object
	Chestplate *Object
	Boots      *Object
	Ring       *Object
}

type Action struct {
	Damage   int
	Sentence string
}

type Stats struct {
	Health    int
	MaxHealth int
	Speed     int
	Defense   int
	Luck      int
}

type Group struct {
	Id      int
	Players map[string]*Player
}

type Behaviour string
type Stat string
type Category string

const (
	Weapon      Category = "Weapon"
	Helmet      Category = "Helmet"
	Chestplate  Category = "Chestplate"
	Boots       Category = "Boots"
	Ring        Category = "Ring"
	Consumable  Category = "Consumable"
	Collectible Category = "Collectible"
)

const (
	Aggressive Behaviour = "Aggressive"
	QuestGiver Behaviour = "QuestGiver"
	Passive    Behaviour = "Passive"
)

const (
	Attack  Stat = "Attack"
	Defense Stat = "Defense"
	Speed   Stat = "Speed"
	Luck    Stat = "Luck"
)

//instead of a chat struct, a broadcast method so that
// world group and room handle messages thorugh it
