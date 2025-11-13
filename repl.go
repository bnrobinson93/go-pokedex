package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bnrobinson93/go-pokedex/internal/pokecache"
)

type config struct {
	Cache    pokecache.Cache
	Count    int    `json:count`
	Previous string `json:previous`
	Next     string `json:next`
	Results  []struct {
		Name string `json:name`
		URL  string `json:url`
	} `json:results`
}

type cliCommand struct {
	name        string
	description string
	callback    func(*config) error
}

var commands = make(map[string]cliCommand)

/*
CleanInput splits the user's input into words.
Note: it will also lowercase and trim
*/
func CleanInput(text string) []string {
	var output []string
	output = append(output, strings.Fields(strings.ToLower(strings.TrimSpace(text)))...)

	return output
}

func REPL(scanner *bufio.Scanner) {
	cache := pokecache.NewCache(5 * time.Second)

	cfg := &config{
		Next:     "https://pokeapi.co/api/v2/location-area?offset=0&limit=20",
		Previous: "",
		Cache:    cache,
	}

	commands = map[string]cliCommand{
		"exit": {
			name:        "exit",
			description: "Exit the application",
			callback:    commandExit,
		},
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
		},
		"map": {
			name:        "map",
			description: "Displays the names of the next 20 location areas in the Pokemon world",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Displays the names of the previous 20 location areas in the Pokemon world",
			callback:    commandMapBackwards,
		},
	}

	for {
		fmt.Print("Pokedex > ")

		if scanner.Scan() {
			words := CleanInput(scanner.Text())
			firstWord := words[0]

			command, exists := commands[firstWord]
			if !exists {
				fmt.Println("Unknown command")
			}

			if err := command.callback(cfg); err != nil {
				fmt.Println(err)
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Printf("Error reading stdin: %v", err)
			return
		}
	}
}

func commandExit(_ *config) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(_ *config) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Printf("Usage:\n\n")

	for _, v := range commands {
		fmt.Printf("  %s: %s\n", v.name, v.description)
	}
	return nil
}

func commandMap(cfg *config) error {
	nextURL := cfg.Next
	if nextURL == "" {
		return fmt.Errorf("there are no more entries to show")
	}

	var value []byte
	if val, found := cfg.Cache.Get(nextURL); found {
		value = val
	} else {
		req, err := http.NewRequest("GET", nextURL, nil)
		if err != nil {
			return fmt.Errorf("unable to generate request: %w", err)
		}

		client := http.Client{}
		res, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("network Error: %w", err)
		}
		defer res.Body.Close()

		val, err := io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("unable to read response body: %w", err)
		}
		value = val
	}

	var data config
	if err := json.Unmarshal(value, &data); err != nil {
		return fmt.Errorf("unable to decode response: %w", err)
	}

	cfg.Next = data.Next
	cfg.Previous = data.Previous

	for _, item := range data.Results {
		fmt.Printf("%s\n", item.Name)
	}

	return nil
}

func commandMapBackwards(cfg *config) error {
	prevURL := cfg.Previous
	if prevURL == "" {
		return fmt.Errorf("cannot go further back")
	}

	var value []byte
	if val, found := cfg.Cache.Get(prevURL); found {
		value = val
	} else {
		req, err := http.NewRequest("GET", prevURL, nil)
		if err != nil {
			return fmt.Errorf("unable to generate request: %w", err)
		}

		client := http.Client{}
		res, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("network Error: %w", err)
		}
		defer res.Body.Close()
	}

	var data config
	if err := json.Unmarshal(value, &data); err != nil {
		return fmt.Errorf("unable to decode response: %w", err)
	}

	cfg.Next = data.Next
	cfg.Previous = data.Previous

	for _, item := range data.Results {
		fmt.Printf("%s\n", item.Name)
	}

	return nil
}
