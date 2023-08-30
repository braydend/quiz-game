package quizgame

import (
	"log"

	"github.com/fasthttp/websocket"
	"github.com/google/uuid"
)

type player struct {
	id      string
	name    string
	isReady bool
	score   int
	game    *Game
	conn    *websocket.Conn
}

func newPlayer(game *Game, conn *websocket.Conn) player {
	uuid, err := uuid.NewUUID()

	if err != nil {
		log.Fatalf("Failed to generate UUID")
	}

	id := uuid.String()
	return player{id, id, false, 0, game, conn}
}

func (p *player) toggleReady() {
	p.isReady = !p.isReady
}

func (p *player) increaseScore() {
	p.score += 1
}

func (p *player) setName(name string) {
	p.name = name
}

/*
Send a message to a specific client
*/
func (player *player) directMessage(msg []byte) {
	if err := player.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
		log.Println("write:", err)
	}
}
