// worldcheck carica il file del mondo e stampa cosa contiene ogni stanza.
// Serve solo per provare il caricatore: non fa parte del server.
//
// Uso: go run ./cmd/worldcheck -world world.json
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"

	"tap/internal/world"
)

func main() {
	path := flag.String("world", "world.json", "world file to load")
	flag.Parse()

	w, err := world.Load(*path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("loaded %d rooms, %d item types, %d npcs, %d quests\n",
		len(w.Rooms), len(w.Items), len(w.Npcs), len(w.Quests))
	fmt.Printf("start room: %s, respawn room: %s\n\n", w.StartRoom.Id, w.RespawnRoom.Id)

	problems := 0
	seen := make(map[int]string) // ItemNumber -> room where it was found

	for _, rid := range sortedRoomIDs(w) {
		room := w.Rooms[rid]
		fmt.Printf("[%s] %s\n", room.Id, room.Name)

		// uscite, in ordine alfabetico per avere sempre lo stesso output
		dirs := make([]string, 0, len(room.AccessibleRooms))
		for dir := range room.AccessibleRooms {
			dirs = append(dirs, dir)
		}
		sort.Strings(dirs)
		for _, dir := range dirs {
			fmt.Printf("    exit %-6s -> %s\n", dir, room.AccessibleRooms[dir].Id)
		}

		for nid, npc := range room.Npc {
			fmt.Printf("    npc  %s (%s, hp %d, alive %v)\n", nid, npc.Behaviour, npc.Hp, npc.IsAlive)
		}

		// oggetti, ordinati per numero di istanza
		nums := make([]int, 0, len(room.Objects))
		for n := range room.Objects {
			nums = append(nums, n)
		}
		sort.Ints(nums)
		for _, n := range nums {
			obj := room.Objects[n]
			fmt.Printf("    item #%d %s (%s)\n", obj.ItemNumber, obj.Id, obj.Name)

			// controlli sulle istanze
			if obj.ItemNumber != n {
				fmt.Printf("    !! key %d but ItemNumber %d\n", n, obj.ItemNumber)
				problems++
			}
			if obj.ItemNumber == 0 {
				fmt.Printf("    !! ItemNumber 0: this is a catalog template, not an instance\n")
				problems++
			}
			if prev, dup := seen[obj.ItemNumber]; dup {
				fmt.Printf("    !! ItemNumber %d also found in %s\n", obj.ItemNumber, prev)
				problems++
			}
			seen[obj.ItemNumber] = room.Id
			if obj == w.Items[obj.Id] {
				fmt.Printf("    !! this is the catalog pointer itself, not a copy\n")
				problems++
			}
		}
		fmt.Println()
	}

	// il catalogo non deve essere stato toccato da NewItem
	for id, tmpl := range w.Items {
		if tmpl.ItemNumber != 0 {
			fmt.Printf("!! catalog item %s has ItemNumber %d (should be 0)\n", id, tmpl.ItemNumber)
			problems++
		}
	}

	fmt.Printf("%d item instances, next ItemNumber will be %d\n", len(seen), w.ItemCount)
	if problems > 0 {
		fmt.Printf("%d problems found\n", problems)
		os.Exit(1)
	}
	fmt.Println("no problems found")
}

func sortedRoomIDs(w *world.World) []string {
	ids := make([]string, 0, len(w.Rooms))
	for id := range w.Rooms {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}