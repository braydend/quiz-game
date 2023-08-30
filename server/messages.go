package quizgame

import (
	"fmt"
	"log"
	"strings"

	"github.com/fasthttp/websocket"
)

type Message struct {
	Command string      `json:"command"`
	Payload interface{} `json:"payload"`
}

type CorrectAnswerPayload struct {
	Name   string `json:"name"`
	Sprite string `json:"sprite"`
}

type PromptPayload struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func messageHandler(c *websocket.Conn) {
	var (
		mt  int
		msg []byte
		err error
	)
	player := clients[c]
	for {
		if mt, msg, err = c.ReadMessage(); err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", msg)

		if strings.HasPrefix(string(msg), "SYS") {
			command, payload := parseCommand(msg)
			switch string(command) {
			case SYS_READY:
				handleReady(c)
				break

			case SYS_UPDATE_NAME:
				handleUpdateName(c, payload)
				break

			case SYS_SYNC:
				handleSync(c)
				break

			default:
				log.Printf("Unknown system command: %s", msg)
			}
		} else if selectedAbility != nil {
			handleGuess(c, msg)
		} else {
			signedMsg := fmt.Sprintf("%s: %s", player.name, msg)
			broadcast(mt, []byte(signedMsg))
		}
	}
}
