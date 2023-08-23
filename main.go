package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

var clients map[*websocket.Conn]*player

const SYS_READY = "SYS_READY"
const SYS_NOT_READY = "SYS_NOT_READY"
const SYS_UPDATE_NAME = "SYS_UPDATE_NAME"

func main() {
	clients = make(map[*websocket.Conn]*player)

	app := fiber.New()

	app.Static("/", "./web/dist/")

	app.Use("/ws", func(c *fiber.Ctx) error {
		// IsWebSocketUpgrade returns true if the client
		// requested upgrade to the WebSocket protocol.
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}

		return fiber.ErrUpgradeRequired
	})

	/*
		Joins connection pool
	*/
	app.Get("/ws/join", websocket.New(func(c *websocket.Conn) {
		_, isExistingConnection := clients[c]

		if !isExistingConnection {
			p := newPlayer()
			clients[c] = &p
			handleUpdateName(c, p.name)
		}

		messageHandler(c)
	}))

	app.Listen(":3000")
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

			default:
				log.Printf("Unknown system command: %s", msg)
			}
		} else {
			broadcast(mt, msg, player.id)
		}
	}
}

/*
Send a message to all established clients
*/
func broadcast(mt int, msg []byte, sender string) {
	signedMsg := fmt.Sprintf("%s: %s", sender, msg)
	for c, _ := range clients {
		if err := c.WriteMessage(mt, []byte(signedMsg)); err != nil {
			log.Println("write:", err)
			// break
		}
	}
}

func parseCommand(msg []byte) (command string, payload string) {
	hasPayload := strings.Contains(string(msg), ":")

	if !hasPayload {
		return string(msg), ""
	}

	splits := strings.Split(string(msg), ":")

	return strings.TrimSpace(string(splits[0])), strings.TrimSpace(string(splits[1]))
}

/*
GAME LOGIC
*/

type player struct {
	id      string
	name    string
	isReady bool
	score   int
}

func handleReady(c *websocket.Conn) {
	player := clients[c]
	player.toggleReady()
	resp := SYS_READY
	if !player.isReady {
		resp = SYS_NOT_READY
	}
	if err := c.WriteMessage(websocket.TextMessage, []byte(resp)); err != nil {
		log.Println("write:", err)
	}
}

func handleUpdateName(c *websocket.Conn, newName string) {
	player := clients[c]
	player.setName(newName)
	resp := fmt.Sprintf("%s: %s", SYS_UPDATE_NAME, newName)
	if err := c.WriteMessage(websocket.TextMessage, []byte(resp)); err != nil {
		log.Println("write:", err)
	}
}

func newPlayer() player {
	uuid, err := uuid.NewUUID()

	if err != nil {
		log.Fatalf("Failed to generate UUID")
	}

	id := uuid.String()
	return player{id, id, false, 0}
}

func (p *player) toggleReady() {
	p.isReady = !p.isReady
}

func (p *player) setName(name string) {
	p.name = name
}

func startGame() {
	broadcast(websocket.TextMessage, []byte("GAME STARTING"), "SERVER")
}
