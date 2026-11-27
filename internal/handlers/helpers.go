package handlers

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"tap/internal/world"
)

// general function to remove every instance of a player in the world
func RemovePlayer(w *world.World, player *world.Player) {
	delete(w.Players, player.Name)
	delete(player.CurrentRoom.Players, player.Name)
	if player.Group != nil {
		delete(player.Group.Players, player.Name)
	}
}

// general function to move player to next room
func MovePlayer(player *world.Player, room *world.Room) {
	delete(player.CurrentRoom.Players, player.Name)
	room.Players[player.Name] = player
	player.CurrentRoom = room
}

// FindItem returns the first item in objects matching arg, or nil.
// arg can be the full id ("item.herbs"), the id without the "item."
// prefix ("herbs") or the display name ("Healing Herbs"); case is ignored.
func FindItem(objects map[int]*world.Object, arg string) *world.Object {
	for _, obj := range objects {
		if strings.EqualFold(obj.Id, arg) ||
			strings.EqualFold(strings.TrimPrefix(obj.Id, "item."), arg) ||
			strings.EqualFold(obj.Name, arg) {
			return obj
		}
	}
	return nil
}

func FindNpc(room *world.Room, arg string) *world.Npc {
	for _, npc := range room.Npc {
		if !npc.IsAlive {
			continue
		}
		if strings.EqualFold(npc.Id, arg) ||
			strings.EqualFold(strings.TrimPrefix(npc.Id, "npc."), arg) ||
			strings.EqualFold(npc.Name, arg) {
			return npc
		}
	}
	return nil
}

func FindQuest(objects map[string]world.Quest, arg string) (world.Quest, bool) {
	for _, obj := range objects {
		if strings.EqualFold(obj.Id, arg) ||
			strings.EqualFold(strings.TrimPrefix(obj.Id, "q."), arg) ||
			strings.EqualFold(obj.Name, arg) {
			return obj, true
		}
	}
	return world.Quest{}, false
}

type questResponse struct {
	QuestID     string `json:"quest_id"`
	Description string `json:"description"`
	Reward      string `json:"reward"`
	Status      string `json:"status"`
	Progress    string `json:"progress,omitempty"`
}

func sendQuest(player *world.Player, q world.Quest, status string, progress string) {
	resp := questResponse{
		QuestID:     q.Id,
		Description: q.Description,
		Reward:      q.Reward,
		Status:      status,
		Progress:    progress,
	}
	data, err := json.Marshal(resp)
	if err != nil {
		fmt.Fprintln(os.Stderr, "quest: marshal:", err)
		Send(player, "ERR 500 INTERNAL_ERROR")
		return
	}
	Send(player, "OK "+string(data))
}

type questListInstance struct {
	QuestID  string `json:"quest_id"`
	Status   string `json:"status"`
	Progress string `json:"progress,omitempty"`
}

func sendQuestList(player *world.Player) {
	questList := make([]questListInstance, 0, len(player.Quests))
	for _, q := range player.Quests {
		sta := "active"
		if q.Completed {
			sta = "completed"
		}
		resp := questListInstance{
			QuestID:  q.Id,
			Status:   sta,
			Progress: fmt.Sprintf("%d/%d", questProgress(player, q), q.Count),
		}
		questList = append(questList, resp)
	}
	data, err := json.Marshal(questList)
	if err != nil {
		fmt.Fprintln(os.Stderr, "quest: marshal:", err)
		Send(player, "ERR 500 INTERNAL_ERROR")
		return
	}
	Send(player, "OK "+string(data))
}

// questProgress returns how far the player is: items owned (fetch) or kills (kill).
func questProgress(player *world.Player, q world.Quest) int {
	if q.Type == world.Fetch {
		return countItems(player.Inventory, q.Target)
	}
	return q.Progress
}

func countItems(inv map[int]*world.Object, typeID string) int {
	n := 0
	for _, obj := range inv {
		if obj.Id == typeID {
			n++
		}
	}
	return n
}

func removeItems(inv map[int]*world.Object, typeID string, n int) {
	for num, obj := range inv {
		if n == 0 {
			return
		}
		if obj.Id == typeID {
			delete(inv, num)
			n--
		}
	}
}

func EquippedItems(player *world.Player) map[int]*world.Object {
	items := make(map[int]*world.Object)
	for _, obj := range player.Equipment {
		items[obj.ItemNumber] = obj
	}
	return items
}

// PlayerStats returns the player's base stats plus the bonuses of the
// equipped items. It never modifies player.Stats.
func PlayerStats(player *world.Player) world.Stats {
	s := player.Stats // copy
	for _, item := range player.Equipment {
		switch item.Category {
		case world.Weapon:
			s.Attack += item.Stat
		case world.Helmet, world.Chestplate:
			s.Defense += item.Stat
		case world.Boots:
			s.Speed += item.Stat
		case world.Ring:
			s.Luck += item.Stat
		}
	}
	return s
}
