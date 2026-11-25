package world

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// The types below describe the JSON world file. They are only used while
// loading: Load converts them into the runtime structs defined in structs.go.

type fileWorld struct {
	StartRoom   string              `json:"start_room"`
	RespawnRoom string              `json:"respawn_room"`
	Rooms       map[string]fileRoom `json:"rooms"`
	Items       map[string]fileItem `json:"items"`
	Npcs        map[string]fileNpc  `json:"npcs"`
	Quests      map[string]fileQst  `json:"quests"`
}

type fileRoom struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Width       int               `json:"width"`
	Height      int               `json:"height"`
	Exits       map[string]string `json:"exits"` // direction -> room id
	Items       []string          `json:"items"` // item ids placed in this room
	Npcs        []string          `json:"npcs"`  // npc ids placed in this room
}

type fileItem struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Stat        int    `json:"stat"`
	Obtainable  bool   `json:"obtainable"`
}

type fileNpc struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Behaviour   string       `json:"behaviour"`
	Hp          int          `json:"hp"`
	Dialogs     []string     `json:"dialogs"`
	Actions     []fileAction `json:"actions"`
	Quests      []string     `json:"quests"` // quest ids given by this npc
}

type fileAction struct {
	Damage   int    `json:"damage"`
	Sentence string `json:"sentence"`
}

type fileQst struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Instructions []string `json:"instructions"` // one instruction per step
	Reward       string   `json:"reward"`       // item id, optional
	Unique       bool     `json:"unique"`
}

var validCategories = map[Category]bool{
	Weapon: true, Helmet: true, Chestplate: true, Boots: true,
	Ring: true, Consumable: true, Collectible: true,
}

var validBehaviours = map[Behaviour]bool{
	Aggressive: true, QuestGiver: true, Passive: true,
}

// Load reads the JSON world file at path, validates it and returns the
// populated World. All problems found are reported together in the error.
func Load(path string) (*World, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields() // a typo in a key is an error, not a silent zero value
	var fw fileWorld
	if err := dec.Decode(&fw); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if errs := validate(&fw); len(errs) > 0 {
		return nil, fmt.Errorf("%s: invalid world:\n  %s", path, strings.Join(errs, "\n  "))
	}
	return build(&fw), nil
}

// validate checks every reference in the file and returns one message per problem.
func validate(fw *fileWorld) []string {
	var errs []string
	add := func(format string, a ...any) { errs = append(errs, fmt.Sprintf(format, a...)) }

	if len(fw.Rooms) == 0 {
		add("no rooms defined")
	}
	if _, ok := fw.Rooms[fw.StartRoom]; !ok {
		add("start_room %q does not exist", fw.StartRoom)
	}
	if _, ok := fw.Rooms[fw.RespawnRoom]; !ok {
		add("respawn_room %q does not exist", fw.RespawnRoom)
	}

	for _, id := range sortedKeys(fw.Items) {
		it := fw.Items[id]
		if it.Name == "" {
			add("item %q: missing name", id)
		}
		if !validCategories[Category(it.Category)] {
			add("item %q: unknown category %q", id, it.Category)
		}
	}

	for _, id := range sortedKeys(fw.Quests) {
		q := fw.Quests[id]
		if q.Name == "" {
			add("quest %q: missing name", id)
		}
		if len(q.Instructions) == 0 {
			add("quest %q: needs at least one instruction", id)
		}
		if q.Reward != "" {
			if _, ok := fw.Items[q.Reward]; !ok {
				add("quest %q: reward item %q does not exist", id, q.Reward)
			}
		}
	}

	for _, id := range sortedKeys(fw.Npcs) {
		n := fw.Npcs[id]
		if n.Name == "" {
			add("npc %q: missing name", id)
		}
		if !validBehaviours[Behaviour(n.Behaviour)] {
			add("npc %q: unknown behaviour %q", id, n.Behaviour)
		}
		if n.Hp <= 0 {
			add("npc %q: hp must be > 0", id)
		}
		for _, qid := range n.Quests {
			if _, ok := fw.Quests[qid]; !ok {
				add("npc %q: quest %q does not exist", id, qid)
			}
		}
	}

	// every item / npc is a unique instance: it can be placed in one room only
	itemRoom := map[string]string{}
	npcRoom := map[string]string{}
	for _, rid := range sortedKeys(fw.Rooms) {
		r := fw.Rooms[rid]
		if r.Name == "" {
			add("room %q: missing name", rid)
		}
		for dir, target := range r.Exits {
			if strings.TrimSpace(dir) == "" || strings.ContainsAny(dir, " \t") {
				add("room %q: invalid exit direction %q", rid, dir)
			}
			if _, ok := fw.Rooms[target]; !ok {
				add("room %q: exit %q leads to unknown room %q", rid, dir, target)
			}
		}
		for _, iid := range r.Items {
			if _, ok := fw.Items[iid]; !ok {
				add("room %q: unknown item %q", rid, iid)
			} else if prev, dup := itemRoom[iid]; dup {
				add("item %q placed in both %q and %q", iid, prev, rid)
			} else {
				itemRoom[iid] = rid
			}
		}
		for _, nid := range r.Npcs {
			if _, ok := fw.Npcs[nid]; !ok {
				add("room %q: unknown npc %q", rid, nid)
			} else if prev, dup := npcRoom[nid]; dup {
				add("npc %q placed in both %q and %q", nid, prev, rid)
			} else {
				npcRoom[nid] = rid
			}
		}
	}

	// every room must be reachable from the start room
	if _, ok := fw.Rooms[fw.StartRoom]; ok && len(errs) == 0 {
		seen := map[string]bool{fw.StartRoom: true}
		queue := []string{fw.StartRoom}
		for len(queue) > 0 {
			cur := queue[0]
			queue = queue[1:]
			for _, next := range fw.Rooms[cur].Exits {
				if !seen[next] {
					seen[next] = true
					queue = append(queue, next)
				}
			}
		}
		for _, rid := range sortedKeys(fw.Rooms) {
			if !seen[rid] {
				add("room %q is not reachable from start_room %q", rid, fw.StartRoom)
			}
		}
	}
	return errs
}

// build converts the validated file data into the runtime structs.
func build(fw *fileWorld) *World {
	w := &World{
		Rooms:   make(map[string]*Room, len(fw.Rooms)),
		Players: make(map[string]*Player),
		Items:   make(map[string]*Object, len(fw.Items)),
		Npcs:    make(map[string]*Npc, len(fw.Npcs)),
		Quests:  make(map[string]*Quest, len(fw.Quests)),
	}

	for id, it := range fw.Items {
		w.Items[id] = &Object{
			Id:          id,
			Name:        it.Name,
			Obtainable:  it.Obtainable,
			Category:    Category(it.Category),
			Stat:        it.Stat,
			Description: it.Description,
		}
	}

	for id, q := range fw.Quests {
		quest := &Quest{
			Id:           id,
			Name:         q.Name,
			Description:  q.Description,
			CompleteStep: len(q.Instructions),
			Instruction:  make(map[int]string, len(q.Instructions)),
			Progression:  make(map[int]*Player),
			Unique:       q.Unique,
		}
		for i, s := range q.Instructions {
			quest.Instruction[i] = s
		}
		if q.Reward != "" {
			quest.Reward = *w.Items[q.Reward]
		}
		w.Quests[id] = quest
	}

	for id, n := range fw.Npcs {
		npc := &Npc{
			Id:          id,
			Name:        n.Name,
			Description: n.Description,
			Hp:          n.Hp,
			IsAlice:     true,
			Behaviour:   Behaviour(n.Behaviour),
			Dialogs:     make(map[int]string, len(n.Dialogs)),
			Action:      make(map[int]*Action, len(n.Actions)),
			Quests:      make(map[int]*Quest, len(n.Quests)),
		}
		for i, d := range n.Dialogs {
			npc.Dialogs[i] = d
		}
		for i, a := range n.Actions {
			npc.Action[i] = &Action{Damage: a.Damage, Sentence: a.Sentence}
		}
		for i, qid := range n.Quests {
			npc.Quests[i] = w.Quests[qid]
		}
		w.Npcs[id] = npc
	}

	// first create all rooms, then link exits (a room may point to one created later)
	for id, r := range fw.Rooms {
		room := &Room{
			Id:              id,
			Name:            r.Name,
			Description:     r.Description,
			Width:           r.Width,
			height:          r.Height,
			Npc:             make(map[string]*Npc, len(r.Npcs)),
			Objects:         make(map[string]*Object, len(r.Items)),
			Players:         make(map[string]*Player),
			AccessibleRooms: make(map[string]*Room, len(r.Exits)),
		}
		for _, iid := range r.Items {
			room.Objects[iid] = w.Items[iid]
		}
		for _, nid := range r.Npcs {
			room.Npc[nid] = w.Npcs[nid]
		}
		w.Rooms[id] = room
	}
	for id, r := range fw.Rooms {
		for dir, target := range r.Exits {
			w.Rooms[id].AccessibleRooms[strings.ToLower(dir)] = w.Rooms[target]
		}
	}

	w.StartRoom = w.Rooms[fw.StartRoom]
	w.RespawnRoom = w.Rooms[fw.RespawnRoom]
	return w
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
