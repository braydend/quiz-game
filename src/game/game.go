package game

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/braydend/quiz-game/src/message"
	"github.com/braydend/quiz-game/src/player"
	"github.com/braydend/quiz-game/src/pokeapi"
	"github.com/gofiber/contrib/websocket"
)

type Game struct {
	clients map[*websocket.Conn]*player.Player
	// selectedAbility *pokeapi.AbilityResult
	selectedMove   *pokeapi.MoveResult
	guessedPokemon map[string]bool
}

func NewGame() Game {
	clients := make(map[*websocket.Conn]*player.Player)
	guessedPokemon := make(map[string]bool)

	return Game{clients: clients, guessedPokemon: guessedPokemon}
}

// If user reconnects, check if they have an id cookie so they can resume their place
func (g *Game) ReconnectClient(c *websocket.Conn, userId string) {
	for oldConnection, player := range g.clients {
		if player.Id() == userId {
			g.clients[c] = player
			clients := g.clients
			delete(clients, oldConnection)
			g.messageHandler(c)
		}
	}
}

func (g *Game) AddClient(c *websocket.Conn) {
	_, isExistingConnection := g.clients[c]

	if !isExistingConnection {
		p := player.NewPlayer(c)
		g.clients[c] = &p
		g.syncClient(c)
	}

	g.messageHandler(c)
}

func (g *Game) Start() {
	g.Broadcast([]byte("GAME STARTING"))
	g.guessedPokemon = make(map[string]bool)

	roundTime := 32 * time.Second

	timer := time.NewTimer(roundTime)
	go func() {
		for i := 0; i < 10; i++ {
			g.newRound()
			<-timer.C
			timer.Reset(roundTime)
		}
		g.resetRound()
	}()
}

/*
Send a message to all established clients
*/
func (g *Game) Broadcast(msg []byte) {
	for _, p := range g.clients {
		p.DirectMessage(msg)
	}
}

func (g *Game) messageHandler(c *websocket.Conn) {
	var (
		msg []byte
		err error
	)
	player := g.clients[c]
	for {
		if _, msg, err = c.ReadMessage(); err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", msg)
		var incomingMsg message.Message
		err := json.Unmarshal(msg, &incomingMsg)

		if err != nil {
			log.Printf("Unable to process message: %s", msg)
			break
		}

		if strings.Contains(incomingMsg.Command, "SYS") {
			switch incomingMsg.Command {
			case message.SYS_READY:
				g.handleReady(c, message.GetPayloadAsBool(incomingMsg))
				break

			case message.SYS_UPDATE_NAME:
				g.handleUpdateName(c, message.GetPayloadAsString(incomingMsg))
				break

			case message.SYS_SYNC:
				g.syncClient(c)
				break

			default:
				log.Printf("Unknown system command: %s", msg)
			}
		} else if incomingMsg.Command == "GUESS" {
			g.handleGuess(c, incomingMsg)
		} else {
			signedMsg := fmt.Sprintf("%s: %s", player.Name(), msg)
			g.Broadcast([]byte(signedMsg))
		}
	}
}

func (g *Game) syncClient(c *websocket.Conn) {
	p := g.clients[c]
	p.HandleSync()
	g.broadcastLeaderboard()
	// if g.selectedAbility != nil {
	if g.selectedMove != nil {
		g.broadcastPromptProgress()
	}
	for name, isGuessed := range g.guessedPokemon {
		if isGuessed {
			p.SendCorrectAnswerCommandDirect(name)
		}
	}
}

func (g *Game) broadcastLeaderboard() {
	leaderboard := []message.PlayerDataPayload{}
	for _, player := range g.clients {
		leaderboard = append(leaderboard, message.PlayerDataPayload{Id: player.Id(), Name: player.Name(), Score: player.Score()})
	}

	updateLeaderboardCommand := message.Message{Command: message.SYS_UPDATE_LEADERBOARD, Payload: message.UpdateLeaderboardPayload{
		Scores: leaderboard,
	}}
	g.Broadcast(message.MarshalMessage(updateLeaderboardCommand))
}

func (g *Game) broadcastPromptProgress() {
	remainingAnswersCount := 0

	for _, isGuessed := range g.guessedPokemon {
		if isGuessed {
			remainingAnswersCount += 1
		}
	}

	newPromptProgressCommand := message.Message{
		Command: message.SYS_PROMPT_PROGRESS,
		Payload: message.PromptProgressPayload{
			// TotalAnswers:     len(g.selectedAbility.Pokemon),
			TotalAnswers:     len(g.selectedMove.Pokemon),
			RemainingAnswers: remainingAnswersCount,
		},
	}
	g.Broadcast(message.MarshalMessage(newPromptProgressCommand))
}

func (g *Game) broadcastPrompt() {
	newPromptCommand := message.Message{Command: message.SYS_PROMPT, Payload: g.selectedMove.Name}
	// newPromptCommand := message.Message{Command: message.SYS_PROMPT, Payload: g.selectedAbility.Name}
	g.Broadcast(message.MarshalMessage(newPromptCommand))
}

func (g *Game) broadcastCorrectAnswer(name string) {
	pokemonData := pokeapi.GetPokemonByName(name)
	correctAnswerCommand := message.Message{Command: message.SYS_CORRECT_ANSWER, Payload: message.CorrectAnswerPayload{Name: pokemonData.Name, Sprite: pokemonData.Sprites.FrontDefault}}
	g.Broadcast(message.MarshalMessage(correctAnswerCommand))
}

func (g *Game) newRound() {
	data := pokeapi.GetAllMoves()
	// data := pokeapi.GetAllAbilities()

	i := rand.Intn(len(data.Results))

	randomItem := data.Results[i]

	// ability := pokeapi.GetAbilityByName(randomAbility.Name)
	move := pokeapi.GetMoveByName(randomItem.Name)

	for len(move.Pokemon) < 5 {
		// for len(ability.Pokemon) < 5 {
		randomAbility := data.Results[rand.Intn(len(data.Results))]
		move = pokeapi.GetMoveByName(randomAbility.Name)
		// ability = pokeapi.GetMoveByName(randomAbility.Name)
	}

	g.selectedMove = &move
	// g.selectedAbility = &ability
	g.guessedPokemon = make(map[string]bool)
	for _, pokemon := range move.Pokemon {
		// for _, pokemon := range ability.Pokemon {
		// g.guessedPokemon[pokemon.Pokemon.Name] = false
		g.guessedPokemon[pokemon.Name] = false
	}

	g.broadcastPrompt()
	g.broadcastPromptProgress()
}

func (g *Game) resetRound() {
	g.selectedMove = nil
	// g.selectedAbility = nil
	g.guessedPokemon = make(map[string]bool)

	g.Broadcast(message.MarshalMessage(message.Message{Command: message.SYS_CLEAR_PROMPT}))
}

func (g *Game) handleUpdateName(c *websocket.Conn, newName string) {
	player := g.clients[c]
	player.SetName(newName)
	player.SendUpdateUserDataCommand()
}

func (g *Game) handleReady(c *websocket.Conn, isReady bool) {
	player := g.clients[c]
	player.SetReady(isReady)

	readyCommand := message.Message{Command: message.SYS_READY, Payload: isReady}
	player.DirectMessage(message.MarshalMessage(readyCommand))

	if g.isEveryoneReady() {
		g.Start()
	}
}

func (g *Game) isEveryoneReady() bool {
	isAllReady := true

	for _, player := range g.clients {
		isAllReady = player.IsReady() && isAllReady
	}

	return isAllReady
}

func (g *Game) handleGuess(c *websocket.Conn, msg message.Message) {
	log.Printf("Guess %s", msg)
	var validAnswers []string
	for _, pokemon := range g.selectedMove.Pokemon {
		// for _, pokemon := range g.selectedAbility.Pokemon {
		validAnswers = append(validAnswers, pokemon.Name)
		// validAnswers = append(validAnswers, pokemon.Pokemon.Name)
	}
	log.Printf("Valid answers: %s", strings.Join(validAnswers, ","))
	guess := message.GetPayloadAsString(msg)
	player := g.clients[c]

	// for _, pokemon := range g.selectedAbility.Pokemon {
	for _, pokemon := range g.selectedMove.Pokemon {
		// if strings.ToUpper(pokemon.Pokemon.Name) == strings.ToUpper(guess) {
		if strings.ToUpper(pokemon.Name) == strings.ToUpper(guess) {
			// if isGuessed := g.guessedPokemon[pokemon.Pokemon.Name]; !isGuessed {
			if isGuessed := g.guessedPokemon[pokemon.Name]; !isGuessed {
				player.IncreaseScore()
				g.guessedPokemon[pokemon.Name] = true
				// g.guessedPokemon[pokemon.Pokemon.Name] = true
				player.SendUpdateScoreCommand()
				g.broadcastCorrectAnswer(pokemon.Name)
				// g.broadcastCorrectAnswer(pokemon.Pokemon.Name)
				g.broadcastLeaderboard()
				g.broadcastPromptProgress()
			}
		}
	}
}
