package game

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strings"

	"github.com/braydend/quiz-game/src/message"
	"github.com/braydend/quiz-game/src/player"
	"github.com/braydend/quiz-game/src/pokeapi"
	"github.com/gofiber/contrib/websocket"
)

type Game struct {
	clients         map[*websocket.Conn]*player.Player
	selectedAbility *pokeapi.AbilityResult
	guessedPokemon  map[string]bool
}

func NewGame() Game {
	clients := make(map[*websocket.Conn]*player.Player)
	guessedPokemon := make(map[string]bool)

	return Game{clients: clients, guessedPokemon: guessedPokemon}
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

	g.newRound()
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
	if g.selectedAbility != nil {
		g.broadcastPrompt()
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

func (g *Game) broadcastPrompt() {
	remainingAnswersCount := 0

	for _, isGuessed := range g.guessedPokemon {
		if isGuessed {
			remainingAnswersCount += 1
		}
	}

	newPromptCommand := message.Message{
		Command: message.SYS_PROMPT,
		Payload: message.PromptPayload{
			Prompt:           g.selectedAbility.Name,
			TotalAnswers:     len(g.selectedAbility.Pokemon),
			RemainingAnswers: remainingAnswersCount,
		},
	}
	g.Broadcast(message.MarshalMessage(newPromptCommand))
}

func (g *Game) broadcastCorrectAnswer(name string) {
	pokemonData := pokeapi.GetPokemonByName(name)
	correctAnswerCommand := message.Message{Command: message.SYS_CORRECT_ANSWER, Payload: message.CorrectAnswerPayload{Name: pokemonData.Name, Sprite: pokemonData.Sprites.FrontDefault}}
	g.Broadcast(message.MarshalMessage(correctAnswerCommand))
}

func (g *Game) newRound() {
	data := pokeapi.GetAllAbilities()

	i := rand.Intn(len(data.Results))

	randomAbility := data.Results[i]

	ability := pokeapi.GetAbilityByName(randomAbility.Name)
	g.selectedAbility = &ability
	g.guessedPokemon = make(map[string]bool)
	for _, pokemon := range ability.Pokemon {
		g.guessedPokemon[pokemon.Pokemon.Name] = false
	}

	g.broadcastPrompt()
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

	isAllReady := true

	for _, player := range g.clients {
		isAllReady = player.IsReady() && isAllReady
	}

	if isAllReady {
		g.Start()
	}
}

func (g *Game) handleGuess(c *websocket.Conn, msg message.Message) {
	log.Printf("Guess %s", msg)
	var validAnswers []string
	for _, pokemon := range g.selectedAbility.Pokemon {
		validAnswers = append(validAnswers, pokemon.Pokemon.Name)
	}
	log.Printf("Valid answers: %s", strings.Join(validAnswers, ","))
	guess := message.GetPayloadAsString(msg)
	player := g.clients[c]

	for _, pokemon := range g.selectedAbility.Pokemon {
		if strings.ToUpper(pokemon.Pokemon.Name) == strings.ToUpper(guess) {
			if isGuessed := g.guessedPokemon[pokemon.Pokemon.Name]; !isGuessed {
				player.IncreaseScore()
				g.guessedPokemon[pokemon.Pokemon.Name] = true
				player.SendUpdateScoreCommand()
				g.broadcastCorrectAnswer(pokemon.Pokemon.Name)
				g.broadcastLeaderboard()
				g.broadcastPrompt()
			}
		}
	}
}
