package handlers

import (
	"fmt"
	"net"
	"strings"
	"unicode"
	"unicode/utf8"

	"tap/internal/world"
)

func HandleConnect(conn net.Conn, name string, w *world.World) {

	//validazione username
	if !utf8.ValidString(name) ||
		name == "" ||
		len(strings.Fields(name)) != 1 ||
		strings.IndexFunc(name, unicode.IsControl) >= 0 {

		fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
		return
	}

	//protegiamo il mondo con il lock eprche altri possono usarlo allo stesos tmepo
	w.Mu.Lock()

	//check se lo username siste gia
	if _, exists := w.Players[name]; exists {
		w.Mu.Unlock()
		fmt.Fprintln(conn, "ERR 201 NAME_IN_USE")
		return
	}

	//creamo il plauyer
	player := &world.Player{
		Connection: conn,
		Name:       name,

		Stats: world.Stats{
			Health:  world.PlayerStartingHealth, //LUDO! messo tutto nelle costanti DENTRO gameConstats cosi le possiamo editare visto che non abbiamo anxoria appurato come sarail gioco
			Speed:   world.PlayerStartingSpeed,
			Defense: world.PlayerStartingDefense,
			Luck:    world.PlayerStartingLuck,
		},

		Inventory: make(map[int]*world.Object),
		Quests:    make(map[int]world.Quest),
		Invites:   make(map[int]*world.Group),

		CurrentRoom: w.StartRoom,
	}

	//registmo il player nel modno
	w.Players[name] = player
	w.StartRoom.Players[name] = player

	w.Mu.Unlock()

	//risposta al client
	if _, err := fmt.Fprintln(conn, "OK connected"); err != nil {

		w.Mu.Lock()

		if w.Players[name] == player {
			delete(w.Players, name)
		}

		if w.StartRoom.Players[name] == player {
			delete(w.StartRoom.Players, name)
		}

		w.Mu.Unlock()
		return
	}
}
