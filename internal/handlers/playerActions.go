package handlers

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"tap/internal/world"
)

type lookRoom struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Exits       map[string]string `json:"exits"` // direzione -> id stanza
}

type lookResponse struct {
	Room       lookRoom       `json:"room"`
	Players    []string       `json:"players"`
	Items      []string       `json:"items"`
	ItemCounts map[string]int `json:"item_counts"`
	Npcs       []string       `json:"npcs"`
}

func HandleLook(w *world.World, player *world.Player) {
	room := player.CurrentRoom
	lRoom := lookRoom{
		ID:          room.Id,
		Name:        room.Name,
		Description: room.Description,
		Exits:       make(map[string]string),
	}
	for dir, nextRoom := range room.AccessibleRooms {
		lRoom.Exits[dir] = nextRoom.Id
	}

	players := make([]string, 0, len(room.Players))
	for name := range room.Players {
		players = append(players, name)
	}
	sort.Strings(players)
	npcs := make([]string, 0, len(room.Npc))
	for _, npc := range room.Npc {
		if npc.IsAlive {
			npcs = append(npcs, npc.Id)
		}
	}
	sort.Strings(npcs)
	counts := make(map[string]int)
	for _, obj := range room.Objects {
		counts[obj.Id]++
	}

	items := make([]string, 0, len(counts))
	for item := range counts {
		items = append(items, item)
	}
	sort.Strings(items)
	completeRoom := lookResponse{
		Room:       lRoom,
		Players:    players,
		Items:      items,
		ItemCounts: counts,
		Npcs:       npcs,
	}
	data, err := json.Marshal(completeRoom)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERR 500 INTERNAL_ERROR", err)
		return
	}
	Send(player, "OK "+string(data))
}

// missing flag in combat
// and missing respawn mechanic
func HandleMove(w *world.World, dir string, player *world.Player) {
	next, ok := player.CurrentRoom.AccessibleRooms[strings.ToLower(dir)]
	if !ok {
		Send(player, "ERR 301 NO_EXIT")
		return
	}
	MovePlayer(player, next)
	Send(player, "OK room="+next.Id)
}

func HandleWho(w *world.World, player *world.Player) {
	count := len(w.Players)
	Send(player, "OK players="+strconv.Itoa(count))
}

func HandleTake(w *world.World, player *world.Player, item string) {
	room := player.CurrentRoom
	foundItem := FindItem(room.Objects, item)
	if foundItem == nil {
		Send(player, "ERR 404 ITEM_NOT_FOUND")
		return
	}
	player.Inventory[foundItem.ItemNumber] = foundItem
	delete(room.Objects, foundItem.ItemNumber)
	Send(player, "OK taken="+foundItem.Id)
}

func HandleDrop(w *world.World, player *world.Player, item string) {
	room := player.CurrentRoom
	foundItem := FindItem(player.Inventory, item)
	if foundItem == nil {
		Send(player, "ERR 404 ITEM_NOT_IN_INVENTORY")
		return
	}
	room.Objects[foundItem.ItemNumber] = foundItem
	delete(player.Inventory, foundItem.ItemNumber)
	Send(player, "OK taken="+foundItem.Id)
}

func HandleInventory(w *world.World, player *world.Player) {
	itemz := make([]string, len(player.Inventory))
	for _, item := range player.Inventory {
		itemz = append(itemz, item.Id)
	}
	sort.Strings(itemz)

	data, err := json.Marshal(itemz)
	if err != nil {
		fmt.Fprintln(os.Stderr, "inventory: marshal:", err)
		Send(player, "ERR 500 INTERNAL_ERROR")
		return
	}
	Send(player, "OK "+string(data))
}

func HandleTalk(w *world.World, player *world.Player, npcName string) {
	npc := FindNpc(player.CurrentRoom, npcName)
	if npc == nil {
		Send(player, "ERR 404 NPC_NOT_FOUND")
		return
	}
	if len(npc.Dialogs) == 0 {
		Send(player, npc.Name+"has nothing to say")
		return
	}
	i := player.TalkIndex[npc.Id]
	Send(player, "OK "+npc.Dialogs[i])
	player.TalkIndex[npc.Id] = (i + 1) % len(npc.Dialogs)
}

func HandleStatus(w *world.World, player *world.Player) {
	data, err := json.Marshal(player.Stats)
	if err != nil {
		fmt.Fprintln(os.Stderr, "inventory: marshal:", err)
		Send(player, "ERR 500 INTERNAL_ERROR")
		return
	}
	Send(player, "OK "+string(data))
}
