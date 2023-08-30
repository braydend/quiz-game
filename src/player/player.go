package player

import (
	"log"
	"strconv"

	"github.com/braydend/quiz-game/src/message"
	"github.com/braydend/quiz-game/src/pokeapi"
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

func (p *Player) SendUpdateScoreCommand() {
	updateScoreCommand := message.Message{Command: message.SYS_UPDATE_SCORE, Payload: strconv.Itoa(p.Score())}
	p.DirectMessage(message.MarshalMessage(updateScoreCommand))
}

func (p *Player) HandleSync() {
	p.SendUpdateUserDataCommand()
	p.sendUpdateScoreCommand()
	// sendUpdateLeaderboardCommandBroad()
	// if selectedAbility != nil {
	// 	p.sendNewPromptCommandDirect()
	// }

	// for name, isGuessed := range guessedPokemon {
	// 	if isGuessed {
	// 		p.sendCorrectAnswerCommandDirect(name)
	// 	}
	// }
}

func (p *Player) SendUpdateUserDataCommand() {
	updateUserDataCommand := message.Message{Command: message.SYS_UPDATE_USER_DATA, Payload: message.PlayerDataPayload{Id: p.Id(), Name: p.Name(), Score: p.Score()}}
	p.DirectMessage(message.MarshalMessage(updateUserDataCommand))
}

func (p *Player) sendUpdateScoreCommand() {
	updateScoreCommand := message.Message{Command: message.SYS_UPDATE_SCORE, Payload: strconv.Itoa(p.Score())}
	p.DirectMessage(message.MarshalMessage(updateScoreCommand))
}

func (p *Player) SendCorrectAnswerCommandDirect(name string) {
	pokemonData := pokeapi.GetPokemonByName(name)
	correctAnswerCommand := message.Message{Command: message.SYS_CORRECT_ANSWER, Payload: message.CorrectAnswerPayload{Name: pokemonData.Name, Sprite: pokemonData.Sprites.FrontDefault}}
	p.DirectMessage(message.MarshalMessage(correctAnswerCommand))
}
