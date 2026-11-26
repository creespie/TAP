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

			//handleWho(conn, w) DA FAREE diocnae

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

				//handleGroupCreate(conn, w) DA FAREE diocnae

			case "INVITE":

				if len(groupArgs) != 1 {
					handlers.Send(player, "ERR 400 BAD_REQUEST")
					return
				}

				//handleGroupInvite(conn, groupArgs[0], w) DA FAREE diocnae

			case "JOIN":

				if len(groupArgs) != 0 {
					handlers.Send(player, "ERR 400 BAD_REQUEST")
					return
				}

				//handleGroupJoin(conn, w) DA FAREE diocnae

			case "LEAVE":

				if len(groupArgs) != 0 {
					handlers.Send(player, "ERR 400 BAD_REQUEST")
					return
				}

				//handleGroupLeave(conn, w) DA FAREE diocane

			default:
				handlers.Send(player, "ERR 400 BAD_REQUEST")
			}

		case "TAKE":

			if len(args) == 0 {
				handlers.Send(player, "ERR 400 BAD_REQUEST")
				return
			}

			//itemName := strings.Join(args, " ")
			//handleTake(conn, itemName, w)DA FAREE diocane

		case "DROP":

			if len(args) == 0 {
				handlers.Send(player, "ERR 400 BAD_REQUEST")
				return
			}

			//itemName := strings.Join(args, " ")
			//handleDrop(conn, itemName, w) DA FAREE diocane

		case "INVENTORY":

			if len(args) != 0 {
				handlers.Send(player, "ERR 400 BAD_REQUEST")
				return
			}

			//handleInventory(conn, w) DA FAREE diocane

		case "TALK":

			if len(args) == 0 {
				handlers.Send(player, "ERR 400 BAD_REQUEST")
				return
			}

			//npcName := strings.Join(args, " ")
			//handleTalk(conn, npcName, w) DA FAREE diocane

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

			//handleStatus(conn, w) DA FAREE diocane

		case "QUEST":

			if len(args) == 0 {
				handlers.Send(player, "ERR 400 BAD_REQUEST")
				return
			}

			//npcName := strings.Join(args, " ")
			//handleQuest(conn, npcName, w) DA FAREE diocane

		case "QUESTS":

			if len(args) != 0 {
				handlers.Send(player, "ERR 400 BAD_REQUEST")
				return
			}

			//handleQuests(conn, w) DA FAREE diocane

		//scoinsciuto il comando in questo caso , o pure ne impetiamo dei nostri.
		default:
			handlers.Send(player, "ERR 400 UNKNOWN_COMMAND")
		}
	}
}
