package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func cleanInput(text string) []string {
	return strings.Fields(strings.ToLower(text))
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		userInput := ""
		fmt.Print("Pokedex > ")
		if scanner.Scan() {
			userInput = scanner.Text()
		}
		words := cleanInput(userInput)

		fmt.Printf("Your command was: %s\n", words[0])
	}
}
