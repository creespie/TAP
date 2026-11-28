package world

import (
	"math/rand"
	"net"
	"sync"
	"time"
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
	Rand        *rand.Rand
	Created     map[string]int
}

func (w *World) NewItem(typeID string) *Object {
	original := w.Items[typeID]
	inst := *original // copia di tutta la struct
	inst.ItemNumber = w.ItemCount
	w.ItemCount++
	w.Created[typeID]++
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
	MaxItems        int       // max objects in the room
	SpawnTypes      []string  // native item types: the ones placed here by the world file
	LastSpawnCheck  time.Time // last time the respawn check ran here
}

type Npc struct {
	Id          string
	Name        string
	Description string
	Hp          int
	MaxHp       int
	Speed       int
	IsAlive     bool
	Behaviour   Behaviour
	Dialogs     map[int]string
	Action      map[int]*Action
	Quest       *Quest
	DiedAt      time.Time // when the npc was killed
}

type Quest struct {
	Id          string
	Name        string
	Description string
	Reward      string // create a new copy and place it in the player inventory
	Type        QuestType
	Target      string
	Count       int
	Progress    int
	Completed   bool
	Unique      bool
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
	Equipment   map[Category]*Object
	Quests      map[string]Quest
	Group       *Group
	Invites     map[string]*Group
	CurrentRoom *Room
	Outbox      chan string
	TalkIndex   map[string]int
	Target      *Npc
}

type Action struct {
	Damage   int
	Sentence string
}

type Stats struct {
	Health    int
	MaxHealth int
	Attack    int
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
type QuestType string

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

const (
	Fetch QuestType = "fetch"
	Kill  QuestType = "kill"
)

//instead of a chat struct, a broadcast method so that
// world group and room handle messages thorugh it
