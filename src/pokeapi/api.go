package pokeapi

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type AbilitiesResult struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type AbilitiesResponse struct {
	Count   int               `json:"count"`
	Results []AbilitiesResult `json:"results"`
}

type MovesResult struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type MovesResponse struct {
	Count   int           `json:"count"`
	Results []MovesResult `json:"results"`
}

func GetAllMoves() MovesResponse {
	resp, err := http.Get("https://pokeapi.co/api/v2/move?limit=2000")

	if err != nil {
		log.Printf("Unable to lookup moves. %s\n", err)
	}

	var data MovesResponse

	respBytes, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Printf("Unable to parse moves response. %s\n", err)
	}

	err = json.Unmarshal(respBytes, &data)
	if err != nil {
		log.Printf("Unable to parse moves. %s\n", err)
	}

	return data
}

func GetAllAbilities() AbilitiesResponse {
	resp, err := http.Get("https://pokeapi.co/api/v2/ability?limit=500")

	if err != nil {
		log.Printf("Unable to lookup abilities. %s\n", err)
	}

	var data AbilitiesResponse

	respBytes, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Printf("Unable to parse abilities response. %s\n", err)
	}

	err = json.Unmarshal(respBytes, &data)
	if err != nil {
		log.Printf("Unable to parse abilities. %s\n", err)
	}

	return data
}

type PokemonAbility struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type AbilityData struct {
	Pokemon PokemonAbility `json:"pokemon"`
}

type AbilityResult struct {
	Id      string        `json:"id"`
	Name    string        `json:"name"`
	Pokemon []AbilityData `json:"pokemon"`
}

func GetAbilityByName(name string) AbilityResult {
	resp, err := http.Get(fmt.Sprintf("https://pokeapi.co/api/v2/ability/%s", name))

	if err != nil {
		log.Printf("Unable to lookup ability (%s). %s\n", name, err)
	}

	var data AbilityResult

	respBytes, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Printf("Unable to parse ability (%s) response. %s\n", name, err)
	}

	err = json.Unmarshal(respBytes, &data)
	if err != nil {
		log.Printf("Unable to parse ability (%s). %s\n", name, err)
	}

	return data
}

type PokemonMove struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type MoveResult struct {
	Id      string        `json:"id"`
	Name    string        `json:"name"`
	Pokemon []PokemonMove `json:"learned_by_pokemon"`
}

func GetMoveByName(name string) MoveResult {
	resp, err := http.Get(fmt.Sprintf("https://pokeapi.co/api/v2/move/%s", name))

	if err != nil {
		log.Printf("Unable to lookup move (%s). %s\n", name, err)
	}

	var data MoveResult

	respBytes, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Printf("Unable to parse move (%s) response. %s\n", name, err)
	}

	err = json.Unmarshal(respBytes, &data)
	if err != nil {
		log.Printf("Unable to parse move (%s). %s\n", name, err)
	}

	return data
}

type PokemonSprite struct {
	FrontDefault string `json:"front_default"`
}

type PokemonResult struct {
	Id      int           `json:"id"`
	Name    string        `json:"name"`
	Sprites PokemonSprite `json:"sprites"`
}

func GetPokemonByName(name string) PokemonResult {
	resp, err := http.Get(fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%s", name))

	if err != nil {
		log.Printf("Unable to lookup pokemon (%s). %s\n", name, err)
	}

	var data PokemonResult

	respBytes, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Printf("Unable to parse pokemon (%s) response. %s\n", name, err)
	}

	err = json.Unmarshal(respBytes, &data)
	if err != nil {
		log.Printf("Unable to parse pokemon (%s). %s\n", name, err)
	}

	return data
}
