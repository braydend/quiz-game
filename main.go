package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strings"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// TODO: Move to game struct
var clients map[*websocket.Conn]*player
var selectedAbility *AbilityResult
var guessedPokemon map[string]bool

const SYS_READY = "SYS_READY"
const SYS_NOT_READY = "SYS_NOT_READY"
const SYS_UPDATE_NAME = "SYS_UPDATE_NAME"
const SYS_CORRECT_ANSWER = "SYS_CORRECT_ANSWER"
const SYS_SYNC = "SYS_SYNC"
const SYS_UPDATE_SCORE = "SYS_UPDATE_SCORE"

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
		TODO: put a room/game ID in the url as a param
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

	app.Listen("0.0.0.0:3000")
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

/*
Send a message to a specific client
*/
func directMessage(c *websocket.Conn, mt int, msg []byte) {
	if err := c.WriteMessage(mt, msg); err != nil {
		log.Println("write:", err)
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

func handleSync(c *websocket.Conn) {
	player := clients[c]
	directMessage(c, websocket.TextMessage, []byte(fmt.Sprintf("%s:%s", SYS_UPDATE_NAME, player.name)))
	directMessage(c, websocket.TextMessage, []byte(fmt.Sprintf("%s:%d", SYS_UPDATE_SCORE, player.score)))
	for name, isGuessed := range guessedPokemon {
		if isGuessed {
			directMessage(c, websocket.TextMessage, []byte(fmt.Sprintf("%s:%s", SYS_CORRECT_ANSWER, name)))
		}
	}
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

func handleGuess(c *websocket.Conn, msg []byte) {
	log.Printf("Guess %s", msg)
	var validAnswers []string
	for _, pokemon := range selectedAbility.Pokemon {
		validAnswers = append(validAnswers, pokemon.Pokemon.Name)
	}
	log.Printf("Valid answers: %s", strings.Join(validAnswers, ","))
	guess := string(msg)
	player := clients[c]

	for _, pokemon := range selectedAbility.Pokemon {
		if strings.ToUpper(pokemon.Pokemon.Name) == strings.ToUpper(guess) {
			if isGuessed := guessedPokemon[pokemon.Pokemon.Name]; !isGuessed {
				player.increaseScore()
				guessedPokemon[pokemon.Pokemon.Name] = true
				broadcast(websocket.TextMessage, []byte(fmt.Sprintf("%s guessed correctly! Their score is now: %d", player.name, player.score)))
				broadcast(websocket.TextMessage, []byte(fmt.Sprintf("%s:%s", SYS_CORRECT_ANSWER, guess)))
				directMessage(c, websocket.TextMessage, []byte(fmt.Sprintf("%s:%d", SYS_UPDATE_SCORE, player.score)))
			}
		}
	}
}

func handleReady(c *websocket.Conn) {
	player := clients[c]
	player.toggleReady()
	resp := SYS_READY
	if !player.isReady {
		resp = SYS_NOT_READY
	}
	directMessage(c, websocket.TextMessage, []byte(resp))

	//TODO: remove
	startGame()
}

func handleUpdateName(c *websocket.Conn, newName string) {
	player := clients[c]
	player.setName(newName)
	resp := fmt.Sprintf("%s: %s", SYS_UPDATE_NAME, newName)
	directMessage(c, websocket.TextMessage, []byte(resp))
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

func (p *player) increaseScore() {
	p.score += 1
}

func (p *player) setName(name string) {
	p.name = name
}

func startGame() {
	broadcast(websocket.TextMessage, []byte("GAME STARTING"))
	guessedPokemon = make(map[string]bool)

	// TODO: remove
	data := getAbilities()

	i := rand.Intn(len(data.Results))

	randomAbility := data.Results[i]

	ability := getAbility(randomAbility.Name)
	selectedAbility = &ability
	for _, pokemon := range ability.Pokemon {
		guessedPokemon[pokemon.Pokemon.Name] = false
	}

	broadcast(websocket.TextMessage, []byte(ability.Name))

	log.Printf("%v", data)
}

/**
 API LOGIC
**/

type AbilitiesResult struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type AbilitiesResponse struct {
	Count   int               `json:"count"`
	Results []AbilitiesResult `json:"results"`
}

func getAbilities() AbilitiesResponse {
	resp, err := http.Get("https://pokeapi.co/api/v2/ability")

	if err != nil {
		log.Printf("Unable to lookup abilities. %s\n", err)
	}

	var data AbilitiesResponse

	respBytes, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Printf("Unable to parse abilities response. %s\n", err)
	}

	log.Printf("%s", respBytes)

	err = json.Unmarshal(respBytes, &data)
	if err != nil {
		log.Printf("Unable to parse abilities. %s\n", err)
	}

	return data
}

type PokemonAbility struct {
	Name string `json:"name"`
}

type AbilityData struct {
	Pokemon PokemonAbility `json:"pokemon"`
}

type AbilityResult struct {
	Id      string        `json:"id"`
	Name    string        `json:"name"`
	Pokemon []AbilityData `json:"pokemon"`
}

func getAbility(name string) AbilityResult {
	resp, err := http.Get(fmt.Sprintf("https://pokeapi.co/api/v2/ability/%s", name))

	if err != nil {
		log.Printf("Unable to lookup ability (%s). %s\n", name, err)
	}

	var data AbilityResult

	respBytes, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Printf("Unable to parse ability (%s) response. %s\n", name, err)
	}

	log.Printf("%s", respBytes)

	err = json.Unmarshal(respBytes, &data)
	if err != nil {
		log.Printf("Unable to parse ability (%s). %s\n", name, err)
	}

	return data
}
