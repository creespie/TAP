package world

import{
	"sync"
}

type World struct{
	Mu sync.Mutex
	Rooms map[int]Room
	Players map[str]Player
}

type Room struct{
	Mu sync.Mutex
	Npc map[str]Npc
	Objects map[str]Objects
	Width, height int
	Players map[str]Player
	Accessible_rooms map[int]Room
	Npc_position [2]int
	Obj_position [2]int
	Room_chat Chat
}

