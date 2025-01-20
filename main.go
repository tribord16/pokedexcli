package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/tribord16/pokedexcli/internal/pokecache"
)

var api_url = "https://pokeapi.co/api/v2/"
var cache *pokecache.Cache

type cliCommand struct {
	name        string
	description string
	callback    func(*Config, string) error
}

type LocationArea struct {
	Name string `json:"name"`
}

type LocationAreaResponse struct {
	Results  []LocationArea `json:"results"`
	Next     string         `json:"next"`
	Previous string         `json:"previous"`
}

type PokemonEncounter struct {
	Pokemon Pokemon `json:"pokemon"`
}

type PokemonStat struct {
	Name string `json:"name"`
}
type PokemonStats struct {
	Base_Stat int         `json:"base_stats"`
	Stat      PokemonStat `json:"stat"`
}
type PokemonType struct {
	Name string `json:"name"`
}
type PokemonTypes struct {
	Type PokemonType `json:"type"`
}
type Pokemon struct {
	Name   string         `json:"name"`
	Height int            `json:"height"`
	Weight int            `json:"weight"`
	Stats  []PokemonStats `json:"stats"`
	Types  []PokemonTypes `json:"types"`
}
type ExploreAreaResponse struct {
	PokemonEncounters []PokemonEncounter `json:"pokemon_encounters"`
}

type PokemonInfoResponse struct {
	Base_experience int            `json:"base_experience"`
	Name            string         `json:"name"`
	Height          int            `json:"height"`
	Weight          int            `json:"weight"`
	Stats           []PokemonStats `json:"stats"`
	Types           []PokemonTypes `json:"types"`
}
type Config struct {
	Next     string
	Previous string
}

var mapCmd map[string]cliCommand
var pokedex map[string]Pokemon

func cleanInput(text string) []string {
	return strings.Fields(strings.ToLower(text))
}

func getLocationAreas(url string) (LocationAreaResponse, error) {
	res, err := http.Get(url)
	if err != nil {
		return LocationAreaResponse{}, fmt.Errorf("network error: %w", err)
	}
	defer res.Body.Close()

	var response LocationAreaResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return LocationAreaResponse{}, fmt.Errorf("error decoding response: %w", err)
	}

	return response, nil
}

func getPokemonArea(url string, area string) (ExploreAreaResponse, error) {
	res, err := http.Get(url + "location-area/" + "/" + area)
	if err != nil {
		return ExploreAreaResponse{}, fmt.Errorf("network error: %w", err)
	}
	defer res.Body.Close()

	var response ExploreAreaResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return ExploreAreaResponse{}, fmt.Errorf("error decoding response: %w", err)
	}

	return response, nil
}

func getPokemonInfo(url string, pokemonName string) (PokemonInfoResponse, error) {
	res, err := http.Get(url + "pokemon/" + pokemonName)
	if err != nil {
		return PokemonInfoResponse{}, fmt.Errorf("network error: %w", err)
	}

	defer res.Body.Close()

	var response PokemonInfoResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return PokemonInfoResponse{}, fmt.Errorf("error decoding response: %w", err)
	}
	rawData, _ := json.Marshal(response)
	fmt.Println("Raw API Response:", string(rawData))
	return response, nil
}

func commandExit(c *Config, areaName string) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandMap(c *Config, areaName string) error {
	if c.Next == "" {
		c.Next = "https://pokeapi.co/api/v2/location-area/?limit=20"
	}

	result, ok := cache.Get(c.Next)
	if ok {
		var locationData LocationAreaResponse
		err := json.Unmarshal(result, &locationData)
		if err != nil {
			return err
		}

		for _, location := range locationData.Results {
			fmt.Println(location.Name)
		}
	} else {
		response, err := getLocationAreas(c.Next)
		if err != nil {
			return err
		}

		data, err := json.Marshal(response)
		if err != nil {
			return err
		}

		cache.Add(c.Next, data)

		for _, location := range response.Results {
			fmt.Println(location.Name)
		}

		c.Next = response.Next
		c.Previous = response.Previous
	}

	if c.Next == "" {
		fmt.Println("No more location areas available")
	}
	return nil
}

func commandMapb(c *Config, areaName string) error {
	if c.Previous == "" {
		fmt.Println("You're on the first page.")
		return nil
	}

	result, ok := cache.Get(c.Previous)
	if ok {
		var locationData LocationAreaResponse
		err := json.Unmarshal(result, &locationData)
		if err != nil {
			return err
		}

		for _, location := range locationData.Results {
			fmt.Println(location.Name)
		}
	} else {
		response, err := getLocationAreas(c.Previous)
		if err != nil {
			return err
		}

		data, err := json.Marshal(response)
		if err != nil {
			return err
		}

		cache.Add(c.Previous, data)

		for _, location := range response.Results {
			fmt.Println(location.Name)
		}

		c.Next = response.Next
		c.Previous = response.Previous
	}
	return nil
}

func commandCatch(c *Config, pokemonName string) error {
	fmt.Printf("Throwing a Pokeball at %s...\n", pokemonName)

	response, err := getPokemonInfo(api_url, pokemonName)
	if err != nil {
		return err
	}
	const difficulte = 100
	const pente = 30

	probability := float64(difficulte) / (float64(response.Base_experience) + float64(pente))
	catchProbability := rand.Float64()
	if catchProbability <= probability {
		fmt.Printf("%s was caught!\n", pokemonName)

		// Add the Pokémon to the Pokedex (if not already caught)
		if _, exists := pokedex[pokemonName]; !exists {
			pokedex[pokemonName] = Pokemon{
				Name:   pokemonName,
				Height: response.Height,
				Weight: response.Weight,
				Stats:  response.Stats,
				Types:  response.Types,
			}
		}
	} else {
		// If the Pokémon escapes
		fmt.Printf("%s escaped!\n", pokemonName)
	}
	return nil
}

func commandExplore(c *Config, areaName string) error {
	result, ok := cache.Get(areaName)
	if ok {
		var locationPokemon ExploreAreaResponse
		err := json.Unmarshal(result, &locationPokemon)
		if err != nil {
			return err
		}

		for _, pokemon := range locationPokemon.PokemonEncounters {
			fmt.Println(pokemon.Pokemon.Name)
		}
	} else {
		response, err := getPokemonArea(api_url, areaName)
		if err != nil {
			return err
		}

		data, err := json.Marshal(response)
		if err != nil {
			return err
		}
		cache.Add(areaName, data)

		fmt.Println("Found Pokemon:")
		for _, pokemon := range response.PokemonEncounters {
			fmt.Println("		" + pokemon.Pokemon.Name)
		}
	}
	return nil
}

func commandInspect(c *Config, pokemonName string) error {
	pokemon, exists := pokedex[pokemonName]
	if !exists {
		fmt.Println("you have not caught that pokemon")
		return nil
	} else {
		fmt.Printf("Name: %s\n Height: %d\n Weight: %d\n", pokemon.Name, pokemon.Height, pokemon.Weight)
		for _, stat := range pokemon.Stats {
			fmt.Printf("-%s: %d\n", stat.Stat.Name, stat.Base_Stat)
		}
	}
	return nil
}
func commandHelp(c *Config, areaName string) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	for _, cmd := range mapCmd {
		_, err := fmt.Printf("%s: %s\n", cmd.name, cmd.description)
		if err != nil {
			return fmt.Errorf("Error printing command info: %w", err)
		}
	}
	return nil
}

func main() {
	config := &Config{}
	rand.Seed(time.Now().UnixNano())

	mapCmd = map[string]cliCommand{
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
		},
		"map": {
			name:        "map",
			description: "Displays the names of 20 location areas in the Pokemon world.",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Displays the previous 20 location areas.",
			callback:    commandMapb,
		},
		"explore": {
			name:        "explore",
			description: "takes 1 location area argument, return pokemons encounters",
			callback:    commandExplore,
		},
		"catch": {
			name:        "catch",
			description: "takes 1 pokemon name argument, catch based on experience",
			callback:    commandCatch,
		},
		"inspect": {
			name:        "inspect",
			description: "takes 1 pokemon name argument, return pokedex info",
			callback:    commandInspect,
		},
	}

	pokedex = map[string]Pokemon{}

	cache = pokecache.NewCache(5 * time.Minute)

	scanner := bufio.NewScanner(os.Stdin)
	for {
		userInput := ""
		fmt.Print("Pokedex > ")
		if !scanner.Scan() {
			fmt.Println("Error reading input")
			break
		}

		userInput = scanner.Text()
		words := cleanInput(userInput)

		cmd, exists := mapCmd[words[0]]
		if exists {
			if len(words) > 1 {
				err := cmd.callback(config, words[1])
				if err != nil {
					fmt.Printf("Error: %v\n", err)
				}
			} else {
				err := cmd.callback(config, "") // Pass empty string if no second argument
				if err != nil {
					fmt.Printf("Error: %v\n", err)
				}
			}
		} else {
			fmt.Println("Unknown command")
		}
	}
}
