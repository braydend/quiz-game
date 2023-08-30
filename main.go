package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"github.com/braydend/quiz-game/src/player"
	"github.com/braydend/quiz-game/src/pokeapi"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

// TODO: Move to game struct
var clients map[*websocket.Conn]*player.Player
var selectedAbility *pokeapi.AbilityResult
var guessedPokemon map[string]bool

const SYS_READY = "SYS_READY"
const SYS_UPDATE_NAME = "SYS_UPDATE_NAME"
const SYS_CORRECT_ANSWER = "SYS_CORRECT_ANSWER"
const SYS_SYNC = "SYS_SYNC"
const SYS_UPDATE_SCORE = "SYS_UPDATE_SCORE"
const SYS_NEW_PROMPT = "SYS_NEW_PROMPT"
const SYS_UPDATE_LEADERBOARD = "SYS_UPDATE_LEADERBOARD"
const SYS_UPDATE_USER_DATA = "SYS_UPDATE_USER_DATA"

type Message struct {
	Command string      `json:"command"`
	Payload interface{} `json:"payload"`
}

type CorrectAnswerPayload struct {
	Name   string `json:"name"`
	Sprite string `json:"sprite"`
}

type updateLeaderboardPayload struct {
	Scores []playerDataPayload `json:"scores"`
}

type playerDataPayload struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Score int    `json:"score"`
}

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = ":3000"
	} else {
		port = ":" + port
	}

	return port
}

func main() {
	clients = make(map[*websocket.Conn]*player.Player)

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
		TODO: put a room/game ID in the url as a param
	*/
	app.Get("/ws/join", websocket.New(func(c *websocket.Conn) {
		_, isExistingConnection := clients[c]

		if !isExistingConnection {
			p := player.NewPlayer(c)
			clients[c] = &p
			handleSync(c)
		}

		messageHandler(c)
	}))

	app.Listen(getPort())
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
		var message Message
		err := json.Unmarshal(msg, &message)

		if err != nil {
			log.Printf("Unable to process message: %s", msg)
			break
		}

		if strings.Contains(message.Command, "SYS") {
			switch message.Command {
			case SYS_READY:
				handleReady(c, getPayloadAsBool(message))
				break

			case SYS_UPDATE_NAME:
				handleUpdateName(c, getPayloadAsString(message))
				break

			case SYS_SYNC:
				handleSync(c)
				break

			default:
				log.Printf("Unknown system command: %s", msg)
			}
		} else if message.Command == "GUESS" {
			handleGuess(c, message)
		} else {
			signedMsg := fmt.Sprintf("%s: %s", player.Name(), msg)
			broadcast(mt, []byte(signedMsg))
		}
	}
}

/*
Send a message to all established clients
*/
func broadcast(mt int, msg []byte) {
	for c, _ := range clients {
		if err := c.WriteMessage(mt, msg); err != nil {
			log.Println("write:", err)
			// break
		}
	}
}

func getPayloadAsBool(msg Message) bool {
	return msg.Payload.(bool)
}

func getPayloadAsString(msg Message) string {
	return msg.Payload.(string)
}

func sendUpdateLeaderboardCommandBroad() {
	leaderboard := []playerDataPayload{}
	for _, player := range clients {
		leaderboard = append(leaderboard, playerDataPayload{Id: player.Id(), Name: player.Name(), Score: player.Score()})
	}

	updateLeaderboardCommand := Message{Command: SYS_UPDATE_LEADERBOARD, Payload: updateLeaderboardPayload{
		Scores: leaderboard,
	}}
	broadcast(websocket.TextMessage, marshalMessage(updateLeaderboardCommand))
}

func sendUpdateLeaderboardCommandDirect(c *websocket.Conn) {
	leaderboard := []playerDataPayload{}
	for _, player := range clients {
		leaderboard = append(leaderboard, playerDataPayload{Id: player.Id(), Name: player.Name(), Score: player.Score()})
	}

	updateLeaderboardCommand := Message{Command: SYS_UPDATE_LEADERBOARD, Payload: updateLeaderboardPayload{
		Scores: leaderboard,
	}}
	player := clients[c]
	player.DirectMessage(marshalMessage(updateLeaderboardCommand))
}

func sendNewPromptCommandBroad() {
	newPromptCommand := Message{Command: SYS_NEW_PROMPT, Payload: selectedAbility.Name}
	broadcast(websocket.TextMessage, marshalMessage(newPromptCommand))
}

// TODO: Change to UPDATE_PROMPT
func sendNewPromptCommandDirect(c *websocket.Conn) {
	newPromptCommand := Message{Command: SYS_NEW_PROMPT, Payload: selectedAbility.Name}
	player := clients[c]
	player.DirectMessage(marshalMessage(newPromptCommand))
}

func sendUpdateNameCommand(c *websocket.Conn) {
	player := clients[c]
	updateNameCommand := Message{Command: SYS_UPDATE_NAME, Payload: player.Name()}
	player.DirectMessage(marshalMessage(updateNameCommand))
}

func sendUpdateUserDataCommand(c *websocket.Conn) {
	player := clients[c]
	updateUserDataCommand := Message{Command: SYS_UPDATE_USER_DATA, Payload: playerDataPayload{Id: player.Id(), Name: player.Name(), Score: player.Score()}}
	player.DirectMessage(marshalMessage(updateUserDataCommand))
}

func sendUpdateScoreCommand(c *websocket.Conn) {
	player := clients[c]
	updateScoreCommand := Message{Command: SYS_UPDATE_SCORE, Payload: strconv.Itoa(player.Score())}
	player.DirectMessage(marshalMessage(updateScoreCommand))
}

func sendCorrectAnswerCommandDirect(c *websocket.Conn, name string) {
	player := clients[c]
	pokemonData := pokeapi.GetPokemonByName(name)
	correctAnswerCommand := Message{Command: SYS_CORRECT_ANSWER, Payload: CorrectAnswerPayload{Name: pokemonData.Name, Sprite: pokemonData.Sprites.FrontDefault}}
	player.DirectMessage(marshalMessage(correctAnswerCommand))
}

func sendCorrectAnswerCommandBroad(name string) {
	pokemonData := pokeapi.GetPokemonByName(name)
	correctAnswerCommand := Message{Command: SYS_CORRECT_ANSWER, Payload: CorrectAnswerPayload{Name: pokemonData.Name, Sprite: pokemonData.Sprites.FrontDefault}}
	broadcast(websocket.TextMessage, marshalMessage(correctAnswerCommand))
}

func handleSync(c *websocket.Conn) {
	sendUpdateUserDataCommand(c)
	sendUpdateScoreCommand(c)
	sendUpdateLeaderboardCommandBroad()
	if selectedAbility != nil {
		sendNewPromptCommandDirect(c)
	}

	for name, isGuessed := range guessedPokemon {
		if isGuessed {
			sendCorrectAnswerCommandDirect(c, name)
		}
	}
}

/*
GAME LOGIC
*/

func handleGuess(c *websocket.Conn, msg Message) {
	log.Printf("Guess %s", msg)
	var validAnswers []string
	for _, pokemon := range selectedAbility.Pokemon {
		validAnswers = append(validAnswers, pokemon.Pokemon.Name)
	}
	log.Printf("Valid answers: %s", strings.Join(validAnswers, ","))
	guess := getPayloadAsString(msg)
	player := clients[c]

	for _, pokemon := range selectedAbility.Pokemon {
		if strings.ToUpper(pokemon.Pokemon.Name) == strings.ToUpper(guess) {
			if isGuessed := guessedPokemon[pokemon.Pokemon.Name]; !isGuessed {
				player.IncreaseScore()
				guessedPokemon[pokemon.Pokemon.Name] = true
				sendUpdateScoreCommand(c)
				broadcast(websocket.TextMessage, []byte(fmt.Sprintf("%s guessed correctly! Their score is now: %d", player.Name(), player.Score())))
				sendCorrectAnswerCommandBroad(pokemon.Pokemon.Name)
				sendUpdateLeaderboardCommandBroad()
			}
		}
	}
}

func handleReady(c *websocket.Conn, isReady bool) {
	player := clients[c]
	player.SetReady(isReady)

	readyCommand := Message{Command: SYS_READY, Payload: isReady}
	player.DirectMessage(marshalMessage(readyCommand))

	isAllReady := true

	for _, player := range clients {
		isAllReady = player.IsReady() && isAllReady
	}

	if isAllReady {
		startGame()
	}
}

func handleUpdateName(c *websocket.Conn, newName string) {
	player := clients[c]
	player.SetName(newName)

	updateNameCommand := Message{Command: SYS_UPDATE_NAME, Payload: newName}
	player.DirectMessage(marshalMessage(updateNameCommand))
}

func startGame() {
	broadcast(websocket.TextMessage, []byte("GAME STARTING"))
	guessedPokemon = make(map[string]bool)

	// TODO: remove
	selectNewPrompt()
	sendNewPromptCommandBroad()
}

func selectNewPrompt() {
	data := pokeapi.GetAllAbilities()

	i := rand.Intn(len(data.Results))

	randomAbility := data.Results[i]

	ability := pokeapi.GetAbilityByName(randomAbility.Name)
	selectedAbility = &ability
	for _, pokemon := range ability.Pokemon {
		guessedPokemon[pokemon.Pokemon.Name] = false
	}

	sendNewPromptCommandBroad()
}

func marshalMessage(msg Message) []byte {
	command, err := json.Marshal(msg)

	if err != nil {
		log.Printf("Failed to marshal command: %v", msg)
	}

	return command
}
