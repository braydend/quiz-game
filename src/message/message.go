package message

import (
	"encoding/json"
	"log"
)

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

type PlayerDataPayload struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Score int    `json:"score"`
}

type UpdateLeaderboardPayload struct {
	Scores []PlayerDataPayload `json:"scores"`
}

func GetPayloadAsBool(msg Message) bool {
	return msg.Payload.(bool)
}

func GetPayloadAsString(msg Message) string {
	return msg.Payload.(string)
}

func MarshalMessage(msg Message) []byte {
	command, err := json.Marshal(msg)

	if err != nil {
		log.Printf("Failed to marshal command: %v", msg)
	}

	return command
}
