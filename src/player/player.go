package player

import (
	"log"

	"github.com/gofiber/contrib/websocket"
	"github.com/google/uuid"
)

type Player struct {
	id      string
	name    string
	isReady bool
	score   int
	conn    *websocket.Conn
}

func NewPlayer(conn *websocket.Conn) Player {
	uuid, err := uuid.NewUUID()

	if err != nil {
		log.Fatalf("Failed to generate UUID")
	}

	id := uuid.String()
	return Player{id, id, false, 0, conn}
}

func (p *Player) SetReady(isReady bool) {
	p.isReady = isReady
}

func (p *Player) IncreaseScore() {
	p.score += 1
}

func (p *Player) SetName(name string) {
	p.name = name
}

func (p Player) Name() string {
	return p.name
}

func (p Player) Id() string {
	return p.id
}

func (p Player) Score() int {
	return p.score
}

func (p Player) IsReady() bool {
	return p.isReady
}

/*
Send a message to a specific client
*/
func (p *Player) DirectMessage(msg []byte) {
	if err := p.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
		log.Println("write:", err)
	}
}
