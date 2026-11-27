package transport

import (
	"fmt"
	"net"
	"strings"
	"tap/internal/handlers"
	"tap/internal/world"
)

func DispatchRequest(conn net.Conn, line string, w *world.World) {
	w.Mu.Lock()
	defer w.Mu.Unlock()
	fields := strings.Fields(line)

	if len(fields) == 0 {
		return
	}

	command := strings.ToUpper(fields[0])
	args := fields[1:]
	if command == "CONNECT" {
		if len(args) != 1 {
			fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
			return
		} else if FindPlayer(w, conn) != nil {
			fmt.Fprintln(conn, "ERR 409 ALREADY_CONNECTED")
			return
		}

		handlers.HandleConnect(conn, args[0], w)
	} else {
		player := FindPlayer(w, conn)
		if player == nil {
			fmt.Fprintln(conn, "ERR 403 NOT_AUTHENTICATED")
			return
		}
		switch command {
		case "QUIT":

			if len(args) != 0 {
				handlers.Send(player, "ERR 400 BAD_REQUEST")
				return
			}
			handlers.RemovePlayer(w, player)
			handlers.Send(player, "OK bye")
			close(player.Outbox)

		case "LOOK":

			if len(args) != 0 {
				handlers.Send(player, "ERR 400 BAD_REQUEST")
				return
			}
			handlers.HandleLook(w, player)

		case "MOVE":

			if len(args) != 1 {
				handlers.Send(player, "ERR 400 BAD_REQUEST")
				return
			}
			handlers.HandleMove(w, args[0], player)

		case "CHAT":

			if len(args) < 2 {
				handlers.Send(player, "ERR 400 BAD_REQUEST")
				return
			}

			scope := strings.ToUpper(args[0])
			message := strings.Join(args[1:], " ")
			switch scope {

			case "GLOBAL":
				handlers.HandleChatGlobal(player, message, w)

			case "ROOM":
				handlers.HandleChatRoom(player, message, w)

			case "GROUP":
				handlers.HandleChatGroup(player, message, w)

			default:
				handlers.Send(player, "ERR 400 BAD_REQUEST")
			}

		case "WHO":

			if len(args) != 0 {
				handlers.Send(player, "ERR 400 BAD_REQUEST")
				return
			}

			handlers.HandleWho(w, player)

		case "GROUP":

			if len(args) < 1 {
				handlers.Send(player, "ERR 400 BAD_REQUEST")
				return
			}

			subcommand := strings.ToUpper(args[0])
			groupArgs := args[1:]

			switch subcommand {

			case "CREATE":

				if len(groupArgs) != 0 {
					handlers.Send(player, "ERR 400 BAD_REQUEST")
					return
				}

				handlers.HandleGroupCreate(w, player)

			case "INVITE":

				if len(groupArgs) != 1 {
					handlers.Send(player, "ERR 400 BAD_REQUEST")
					return
				}

				handlers.HandleGroupInvite(w, groupArgs[0], player)

			case "JOIN":

				if len(groupArgs) != 1 {
					handlers.Send(player, "ERR 400 BAD_REQUEST")
					return
				}

				handlers.HandleGroupJoin(w, groupArgs[0], player)

			case "LEAVE":

				if len(groupArgs) != 0 {
					handlers.Send(player, "ERR 400 BAD_REQUEST")
					return
				}

				handlers.HandleGroupLeave(w, player)

			default:
				handlers.Send(player, "ERR 400 BAD_REQUEST")
			}

		case "TAKE":

			if len(args) == 0 {
				handlers.Send(player, "ERR 400 BAD_REQUEST")
				return
			}

			itemName := strings.ToLower(strings.Join(args, " "))
			handlers.HandleTake(w, player, itemName)

		case "DROP":

			if len(args) == 0 {
				handlers.Send(player, "ERR 400 BAD_REQUEST")
				return
			}

			itemName := strings.ToLower(strings.Join(args, " "))
			handlers.HandleDrop(w, player, itemName)

		case "INVENTORY":

			if len(args) != 0 {
				handlers.Send(player, "ERR 400 BAD_REQUEST")
				return
			}

			handlers.HandleInventory(w, player)

		case "TALK":

			if len(args) == 0 {
				handlers.Send(player, "ERR 400 BAD_REQUEST")
				return
			}

			npcName := strings.Join(args, " ")
			handlers.HandleTalk(w, player, npcName)

		case "ATTACK":

			if len(args) == 0 {
				handlers.Send(player, "ERR 400 BAD_REQUEST")
				return
			}

			//npcName := strings.Join(args, " ")
			//handleAttack(conn, npcName, w) DA FAREE diocane

		// PLAYER STATUS
		case "STATUS":

			if len(args) != 0 {
				handlers.Send(player, "ERR 400 BAD_REQUEST")
				return
			}

			handlers.HandleStatus(w, player)

		case "QUEST":

			if len(args) == 0 {
				handlers.Send(player, "ERR 400 BAD_REQUEST")
				return
			}

			npcName := strings.Join(args, " ")
			handlers.HandleQuest(w, player, npcName)

		case "QUESTS":

			if len(args) != 0 {
				handlers.Send(player, "ERR 400 BAD_REQUEST")
				return
			}

			handlers.HandleQuests(w, player)

		//scoinsciuto il comando in questo caso , o pure ne impetiamo dei nostri.
		default:
			handlers.Send(player, "ERR 400 UNKNOWN_COMMAND")
		}
	}
}
