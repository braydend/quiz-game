package quizgame

import (
	"log"

	"github.com/fasthttp/websocket"
)

type Game struct {
	Id              string
	SelectedAbility AbilityResult
	Answers         map[string]bool
	Players         map[*websocket.Conn]*player
}

func (game *Game) Play() {

}

/*
Send a message to all established clients
*/
func (game *Game) broadcast(mt int, msg []byte) {
	for c, _ := range game.Players {
		if err := c.WriteMessage(mt, msg); err != nil {
			log.Println("write:", err)
			// break
		}
	}
}
