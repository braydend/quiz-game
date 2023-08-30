package quizgame

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/fasthttp/websocket"
)

const SYS_READY = "SYS_READY"
const SYS_NOT_READY = "SYS_NOT_READY"
const SYS_UPDATE_NAME = "SYS_UPDATE_NAME"
const SYS_CORRECT_ANSWER = "SYS_CORRECT_ANSWER"
const SYS_SYNC = "SYS_SYNC"
const SYS_UPDATE_SCORE = "SYS_UPDATE_SCORE"
const SYS_NEW_PROMPT = "SYS_NEW_PROMPT"

func parseCommand(msg []byte) (command string, payload string) {
	hasPayload := strings.Contains(string(msg), ":")

	if !hasPayload {
		return string(msg), ""
	}

	splits := strings.Split(string(msg), ":")

	return strings.TrimSpace(string(splits[0])), strings.TrimSpace(string(splits[1]))
}

func sendNewPromptCommandBroad() {
	newPromptCommand := Message{Command: SYS_NEW_PROMPT, Payload: selectedAbility.Name}
	newPromptCommandBytes, err := json.Marshal(newPromptCommand)

	if err != nil {
		log.Printf("Failed to marshal update name command")
	}

	broadcast(websocket.TextMessage, newPromptCommandBytes)
}

// TODO: Change to UPDATE_PROMPT
func sendNewPromptCommandDirect(c *websocket.Conn) {
	newPromptCommand := Message{Command: SYS_NEW_PROMPT, Payload: selectedAbility.Name}
	newPromptCommandBytes, err := json.Marshal(newPromptCommand)

	if err != nil {
		log.Printf("Failed to marshal update name command")
	}

	directMessage(c, websocket.TextMessage, newPromptCommandBytes)
}

func sendUpdateNameCommand(c *websocket.Conn) {
	player := clients[c]
	updateNameCommand := Message{Command: SYS_UPDATE_NAME, Payload: player.name}
	updateNameCommandBytes, err := json.Marshal(updateNameCommand)

	if err != nil {
		log.Printf("Failed to marshal update name command")
	}
	directMessage(c, websocket.TextMessage, updateNameCommandBytes)
}

func sendUpdateScoreCommand(c *websocket.Conn) {
	player := clients[c]
	updateScoreCommand := Message{Command: SYS_UPDATE_SCORE, Payload: strconv.Itoa(player.score)}
	updateScoreCommandBytes, err := json.Marshal(updateScoreCommand)

	if err != nil {
		log.Printf("Failed to marshal update score command")
	}

	directMessage(c, websocket.TextMessage, updateScoreCommandBytes)
}

func sendCorrectAnswerCommandDirect(c *websocket.Conn, name string) {
	pokemonData := getPokemon(name)
	correctAnswerCommand := Message{Command: SYS_CORRECT_ANSWER, Payload: CorrectAnswerPayload{Name: pokemonData.Name, Sprite: pokemonData.Sprites.FrontDefault}}
	correctAnswerCommandBytes, err := json.Marshal(correctAnswerCommand)

	if err != nil {
		log.Printf("Failed to marshal guess pokemon command")
	}
	directMessage(c, websocket.TextMessage, correctAnswerCommandBytes)
}

func sendCorrectAnswerCommandBroad(name string) {
	pokemonData := getPokemon(name)
	correctAnswerCommand := Message{Command: SYS_CORRECT_ANSWER, Payload: CorrectAnswerPayload{Name: pokemonData.Name, Sprite: pokemonData.Sprites.FrontDefault}}
	correctAnswerCommandBytes, err := json.Marshal(correctAnswerCommand)

	if err != nil {
		log.Printf("Failed to marshal guess pokemon command")
	}
	broadcast(websocket.TextMessage, correctAnswerCommandBytes)
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
				sendUpdateScoreCommand(c)
				broadcast(websocket.TextMessage, []byte(fmt.Sprintf("%s guessed correctly! Their score is now: %d", player.name, player.score)))
				sendCorrectAnswerCommandBroad(pokemon.Pokemon.Name)
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

func handleSync(c *websocket.Conn) {
	sendUpdateNameCommand(c)
	sendUpdateScoreCommand(c)
	if selectedAbility != nil {
		sendNewPromptCommandDirect(c)
	}

	for name, isGuessed := range guessedPokemon {
		if isGuessed {
			sendCorrectAnswerCommandDirect(c, name)
		}
	}
}

func handleUpdateName(c *websocket.Conn, newName string) {
	player := clients[c]
	player.setName(newName)

	updateNameCommand := Message{Command: SYS_UPDATE_NAME, Payload: newName}
	updateNameCommandBytes, err := json.Marshal(updateNameCommand)

	if err != nil {
		log.Printf("Failed to marshal update name command")
	}
	directMessage(c, websocket.TextMessage, updateNameCommandBytes)
}
