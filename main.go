package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type cliCommand struct {
	name        string
	description string
	callback    func(*Config) error
}

type LocationArea struct {
	Name string `json:"name"`
}

type LocationAreaResponse struct {
	Results  []LocationArea `json:"results"`
	Next     string         `json:"next"`
	Previous string         `json:"previous"`
}

type Config struct {
	Next     string
	Previous string
}

var mapCmd map[string]cliCommand

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

func commandExit(c *Config) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandMap(c *Config) error {
	if c.Next == "" {
		c.Next = "https://pokeapi.co/api/v2/location-area/?limit=20"
	}

	response, err := getLocationAreas(c.Next)
	if err != nil {
		return err
	}

	for _, location := range response.Results {
		fmt.Println(location.Name)
	}

	c.Next = response.Next
	c.Previous = response.Previous

	if c.Next == "" {
		fmt.Println("No more location areas available")
	}
	return nil
}

func commandMapb(c *Config) error {
	if c.Previous == "" {
		fmt.Println("You're on the first page.")
		return nil
	}

	response, err := getLocationAreas(c.Previous)
	if err != nil {
		return err
	}

	for _, location := range response.Results {
		fmt.Println(location.Name)
	}

	c.Next = response.Next
	c.Previous = response.Previous

	return nil
}
func commandHelp(c *Config) error {
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
	}

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
			err := cmd.callback(config)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		} else {
			fmt.Println("Unknown command")
		}
	}
}
