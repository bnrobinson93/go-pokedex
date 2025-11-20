package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bnrobinson93/go-pokedex/internal/pokecache"
)

type config struct {
	Cache    pokecache.Cache
	Count    int    `json:"count"`
	Previous string `json:"previous"`
	Next     string `json:"next"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
	Pokedex map[string]Pokemon
}

type locationDetails struct {
	ID                   uint32 `json:"id"`
	Name                 string `json:"name"`
	GameIndex            uint32 `json:"game_index"`
	EncounterMethodRates []struct {
		EncounterMethod struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"encounter_method"`
		VersionDetails []struct {
			Rate    uint8 `json:"rate"`
			Version struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"version"`
		} `json:"version_details"`
	} `json:"encounter_method_rates"`
	Location struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"location"`
	Names []struct {
		Name     string `json:"name"`
		Language struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"language"`
	} `json:"names"`
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"pokemon"`
		VersionDetails []struct {
			Version struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"version"`
			MaxChance        uint8 `json:"max_chance"`
			EncounterDetails []struct {
				MinLevel        uint8 `json:"min_level"`
				MaxLevel        uint8 `json:"max_level"`
				ConditionValues []struct {
					Name string `json:"name"`
					URL  string `json:"url"`
				} `json:"condition_values"`
				Chance uint8 `json:"chance"`
				Method struct {
					Name string `json:"name"`
					URL  string `json:"url"`
				}
			} `json:"encounter_details"`
		} `json:"version_details"`
	} `json:"pokemon_encounters"`
}

type Pokemon struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	BaseExperience int    `json:"base_experience"`
	Height         int    `json:"height"`
	IsDefault      bool   `json:"is_default"`
	Order          int    `json:"order"`
	Weight         int    `json:"weight"`
	Abilities      []struct {
		IsHidden bool `json:"is_hidden"`
		Slot     int  `json:"slot"`
		Ability  struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"ability"`
	} `json:"abilities"`
	Forms []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"forms"`
	GameIndeces []struct {
		GameIndex int `json:"game_index"`
		Version   struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"version"`
	} `json:"game_indices"`
	HeldItems []struct {
		Item struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		}
		VersionDetails         []struct{} `json:"version_details"`
		LocationAreaEncounters string     `json:"location_area_encounters"`
	} `json:"held_items"`
	Moves []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"moves"`
	Species struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"species"`
	Sprites struct{} `json:"sprites"`
	Cries   struct{} `json:"cries"`
	Stats   []struct {
		BaseStat int `json:"base_stat"`
		Effort   int `json:"effort"`
		Stat     struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"stat"`
	} `json:"stats"`
	Types []struct {
		Slot int `json:"slot"`
		Type struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"type"`
	} `json:"types"`
	PastAbilities []struct{} `json:"past_abilities"`
}

type cliCommand struct {
	name        string
	description string
	callback    func(*config, string) error
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
		Pokedex:  map[string]Pokemon{},
	}

	commands = map[string]cliCommand{
		"catch": {
			name:        "catch",
			description: "Catch a Pokemon by name",
			callback:    commandCatch,
		},
		"exit": {
			name:        "exit",
			description: "Exit the application",
			callback:    commandExit,
		},
		"explore": {
			name:        "explore",
			description: "Display the Pokemon which can be found in a given location area by id or name",
			callback:    commandExplore,
		},
		"inspect": {
			name:        "inspect",
			description: "Inspect a caught Pokemon by name",
			callback:    commandInspect,
		},
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
		},
		"pokedex": {
			name:        "pokedex",
			description: "Lists all caught Pokemon in your Pokedex",
			callback:    commandPokedex,
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

	commandHelp(cfg, "")
	fmt.Println("")
	for {
		fmt.Print("Pokedex > ")

		if scanner.Scan() {
			words := CleanInput(scanner.Text())
			if len(words) == 0 {
				continue
			}
			firstWord := words[0]
			args := words[1:]

			command, exists := commands[firstWord]
			if !exists {
				fmt.Println("Unknown command")
			}

			if err := command.callback(cfg, strings.Join(args, " ")); err != nil {
				fmt.Println(err)
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Printf("Error reading stdin: %v", err)
			return
		}
	}
}

func commandExit(_ *config, _ string) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(_ *config, _ string) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Printf("Usage:\n\n")

	for _, v := range commands {
		fmt.Printf("  %s: %s\n", v.name, v.description)
	}
	return nil
}

func commandMap(cfg *config, _ string) error {
	nextURL := cfg.Next
	if nextURL == "" {
		return fmt.Errorf("there are no more entries to show")
	}

	var value []byte
	if val, found := cfg.Cache.Get(nextURL); found {
		fmt.Println("Cache hit!")
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
		cfg.Cache.Add(nextURL, value)
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

func commandMapBackwards(cfg *config, _ string) error {
	prevURL := cfg.Previous
	if prevURL == "" {
		return fmt.Errorf("cannot go further back")
	}

	var value []byte
	if val, found := cfg.Cache.Get(prevURL); found {
		fmt.Println("Cache hit!")
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

		val, err := io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("unable to read response body: %w", err)
		}
		value = val
		cfg.Cache.Add(prevURL, value)
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

func commandExplore(cfg *config, idOrName string) error {
	if idOrName == "" {
		return fmt.Errorf("please provide a location area id or name to explore")
	}

	url := "https://pokeapi.co/api/v2/location-area/" + idOrName

	fmt.Printf("Exploring location area: %s (%s)\n", idOrName, url)

	var value []byte
	if val, found := cfg.Cache.Get(url); found {
		fmt.Println("Cache hit!")
		value = val
	} else {
		res, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("network Error: %w", err)
		}

		defer res.Body.Close()

		val, err := io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("unable to read response body: %w", err)
		}
		value = val
		cfg.Cache.Add(url, value)
	}

	var data locationDetails
	if err := json.Unmarshal(value, &data); err != nil {
		return fmt.Errorf("unable to decode response: %w", err)
	}

	if len(data.PokemonEncounters) > 0 {
		fmt.Println("Found Pokemon:")
	}
	for _, encounter := range data.PokemonEncounters {
		fmt.Printf("- %s\n", encounter.Pokemon.Name)
	}

	return nil
}

func commandCatch(cfg *config, name string) error {
	if name == "" {
		return fmt.Errorf("please provide a Pokemon name to catch")
	}

	fmt.Printf("Throwing a Pokeball at %s...\n", name)

	url := "https://pokeapi.co/api/v2/pokemon/" + name

	var value []byte
	if val, found := cfg.Cache.Get(url); found {
		fmt.Println("Cache hit!")
		value = val
	} else {
		res, err := http.Get(url)
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

	var pokemonStats Pokemon
	if err := json.Unmarshal(value, &pokemonStats); err != nil {
		return fmt.Errorf("unable to decode response: %w", err)
	}

	experience := pokemonStats.BaseExperience
	chances := float32(rand.Intn(experience)) / float32(experience)

	didCatch := chances > 0.75
	if didCatch {
		cfg.Pokedex[name] = pokemonStats
		fmt.Printf("You have caught %s!\n", name)
	}

	return nil
}

func commandInspect(cfg *config, name string) error {
	pokemon, exists := cfg.Pokedex[name]
	if !exists {
		return fmt.Errorf("you do not have %s in your Pokedex", name)
	}

	fmt.Printf("Name: %s\n", pokemon.Name)
	fmt.Printf("Height: %v\n", pokemon.Height)
	fmt.Printf("Weight: %v\n", pokemon.Weight)

	fmt.Println("Stats:")
	for _, stat := range pokemon.Stats {
		fmt.Printf("  - %s: %v\n", stat.Stat.Name, stat.BaseStat)
	}

	fmt.Println("Types:")
	for _, t := range pokemon.Types {
		fmt.Println("  -", t.Type.Name)
	}

	return nil
}

func commandPokedex(cfg *config, _ string) error {
	if len(cfg.Pokedex) == 0 {
		return fmt.Errorf("your Pokedex is empty!")
	}

	fmt.Println("Your Pokedex:")
	for _, entry := range cfg.Pokedex {
		fmt.Printf(" - %s\n", entry.Name)
	}

	return nil
}
