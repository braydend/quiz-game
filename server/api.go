package quizgame

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

func getAbilities() AbilitiesResponse {
	resp, err := http.Get("https://pokeapi.co/api/v2/ability?limit=500")

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

type PokemonSprite struct {
	FrontDefault string `json:"front_default"`
}

type PokemonResult struct {
	Id      int           `json:"id"`
	Name    string        `json:"name"`
	Sprites PokemonSprite `json:"sprites"`
}

func getPokemon(name string) PokemonResult {
	resp, err := http.Get(fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%s", name))

	if err != nil {
		log.Printf("Unable to lookup pokemon (%s). %s\n", name, err)
	}

	var data PokemonResult

	respBytes, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Printf("Unable to parse pokemon (%s) response. %s\n", name, err)
	}

	log.Printf("%s", respBytes)

	err = json.Unmarshal(respBytes, &data)
	if err != nil {
		log.Printf("Unable to parse pokemon (%s). %s\n", name, err)
	}

	return data
}
