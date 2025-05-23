package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/rashadat1/goPokedex/internal/pokecache"
)


var commandRegistry map[string]cliCommand

type cliCommand struct {
	name           string
	description    string
	callback       func(*config) error
}

type config struct {
	Next           string
	Prev           string
	Cache          *pokecache.Cache
	ExploreArg     string
	CatchArg       string
	Pokedex        map[string]UnmarshaledPokemonInfo
	InspectArg     string
}
// A field or function is exported (visible outside package) only if it starts
// with a capital letter - json.Unmarshal() only sees exported fields
type LocationName struct {
	Name           string `json:"name"`
	Url            string `json:"url"`
}
type UnmarshaledLocationAreas struct {
	Count          int `json:"count"`
	Next           string `json:"next"`
	Previous       string `json:"previous"`
	Results        []LocationName `json:"results"`

}

// Pokemon Encounter in Location-Area Structs
type MethodData struct {
	Name               string `json:"name"`
	Url                string `json:"url"`
}
type EncounterData struct {
	Chance             int `json:"chance"`
	MaxLevel           int `json:"max_level"`
	MinLevel           int `json:"min_level"`
	Method             MethodData `json:"method"`
}
type VersionDetails struct {
	EncounterData      []EncounterData `json:"encounter_details"`	
}
type PokemonIdentity struct {
	Name               string `json:"name"`
	Url                string `json:"url"`
}
type EncounterDetails struct {
	Pokemon            PokemonIdentity `json:"pokemon"`
	VersionDetail      []VersionDetails `json:"version_details"`
}
type UnmarshaledPokemonEncounters struct {
	PokemonEncounters  []EncounterDetails `json:"pokemon_encounters"`
}

// Pokemon Info for Catching Structs
type AbilityData struct {
	Ability           Ability `json:"ability"`
	IsHidden          bool `json:"is_hidden"`
}
type Ability struct {
	Name              string `json:"name"`
	Url               string `json:"url"`
}
type MoveData struct {
	VersionDetails    []MoveVersionDetail `json:"version_group_details"`
	Name              string `json:"name"`
	Url               string `json:"url"`
}
type MoveVersionDetail struct {
	VersionGroup      VersionGroupNameMove `json:"version_group"`
	MoveLearnMethod   MoveLearnMethod `json:"move_learn_method"`
	LevelLearnedAt    int `json:"level_learned_at"`
}
type VersionGroupNameMove struct {
	Name              string `json:"name"`
	Url               string `json:"url"`
}
type MoveLearnMethod struct {
	Name              string `json:"name"` 
}
type StatData struct {
	BaseStat          int `json:"base_stat"`
	Effort            int `json:"effort"`
	Stat              Stat `json:"stat"`
}
type Stat struct {
	Name              string `json:"name"`
	Url               string `json:"url"`
}
type TypeData struct {
	Type              Type `json:"type"`
	SlotNum           int `json:"slot"`
}
type Type struct {
	Name              string `json:"name"`
	Url               string `json:"url"`
}
type UnmarshaledPokemonInfo struct {
	Abilities         []AbilityData `json:"abilities"`
	Moves             []MoveData `json:"moves"`
	BaseExp           int `json:"base_experience"`
	BaseStats         []StatData `json:"stats"`
	Type              []TypeData `json:"types"`
	Height            float32 `json:"height"`
	Weight            float32 `json:"weight"`
	EntryDescr		  string
	BaseHappiness     int
	CaptureRate       int
}
// Pokemon Species Structs
type UnmarshaledPokemonSpecies struct {
	FlavorText        []FlavorText `json:"flavor_text_entries"`
	BaseHappiness     int `json:"base_happiness"`
	CaptureRate       int `json:"capture_rate"`

}
type FlavorText struct {
	EntryDescr        string `json:"flavor_text"`
}
// add struct tags so json decoder can match the Go field with the JSON field
var userPokedex map[string]UnmarshaledPokemonInfo

func main() {
	inputReader := bufio.NewScanner(os.Stdin)
	cache := pokecache.NewCache(20 * time.Second)
	userPokedex := make(map[string]UnmarshaledPokemonInfo)
	configuration := config{
		Next: "https://pokeapi.co/api/v2/location-area?offset=0&limit=20",
		Prev: "",
		Cache: cache,
		ExploreArg: "",
		CatchArg: "",
		InspectArg: "",
		Pokedex: userPokedex,
	}

	commandRegistry = make(map[string]cliCommand)
	commandRegistry["exit"] = cliCommand{
		name:           "exit",
		description:    "Exit the Pokedex",
		callback:       commandExit,
	}
	commandRegistry["help"] = cliCommand{
		name:           "help",
		description:    "Displays a help message",
		callback:       commandHelp,
	}
	commandRegistry["map"] = cliCommand{
		name:           "map",
		description:    "Displays the names of 20 location areas (next)",
		callback:       commandMap,
	}
	commandRegistry["mapb"] = cliCommand{
		name:           "mapb",
		description:    "Displays the names of 20 location areas (previous)",
		callback:       commandMapb,
	}
	commandRegistry["explore"] = cliCommand{
		name:           "explore",
		description:    "Displays available pokemon given a location area",
		callback:       commandExplore,
	}
	commandRegistry["catch"] = cliCommand{
		name:           "catch",
		description:    "Attempts to catch the pokemon provided as argument - adding it to the user's Pokedex",
		callback:       commandCatch,
	}
	commandRegistry["inspect"] = cliCommand{
		name:           "inspect",
		description:    "Displays pokedex data for a captured pokemon",
		callback:       commandInspect,
	}
	commandRegistry["pokedex"] = cliCommand{
		name:           "pokedex",
		description:    "Lists all of the pokemon that the user has caught",
		callback:       commandPokedex,
	}
	for {
		_, err := fmt.Fprint(os.Stdout, "Pokedex > ")
		if err != nil {
			log.Fatal("Error starting REPL loop" + err.Error())
		}
		inputReader.Scan()
		if inputReader.Err() != nil {
			log.Println("Error reading input: " + inputReader.Err().Error())
		}
		rawInput := inputReader.Text()
		cleanedInput := cleanInput(rawInput)
		if len(cleanedInput) >= 1 {
			commandName := cleanedInput[0]
			if commandName == "explore" || commandName == "catch" || commandName == "inspect" {
				if len(cleanedInput) != 2 {
					fmt.Printf("%s command takes 1 argument %d\n were given", commandName, len(cleanedInput) - 1)
					continue
				} else {
					if commandName == "explore" {
						configuration.ExploreArg = cleanedInput[1]
					} else if commandName == "catch" {
						configuration.CatchArg = cleanedInput[1]
					} else if commandName == "inspect" {
						configuration.InspectArg = cleanedInput[1]
					}
				}
			}
			commandData, exists := commandRegistry[commandName]
			if exists {
				err = commandData.callback(&configuration)
				if err != nil {
					fmt.Println("Error from callback: " + " from " + commandName + err.Error())
				}
			} else {
				fmt.Println("Unknown command")
			}
		}
	}
}

func cleanInput(text string) []string {
	// split user input into words based on whitespace
	// lowercase input 
	// trim whitespace 
	textStrings := strings.Split(text, " ")
	var cleanedInput []string
	for _, element := range textStrings {
		if element != "" {
			cleanedInput = append(cleanedInput, strings.Trim(strings.ToLower(element), "\r\n ,"))
		}
	}
	return cleanedInput
}

func commandExit(conf *config) error {
	// callback for exit command
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil	
}

func commandHelp(conf *config) error {
	fmt.Print("Welcome to the Pokedex!\r\nUsage:\r\n\r\n")
	for _, value := range commandRegistry {
		fmt.Println(value.name + ": " + value.description)
	}
	return nil
}

func commandMap(conf *config) error {
	urlCache := conf.Cache
	var body []byte
	if conf.Next == "" {
		fmt.Println("you're on the last page")
		return nil
	}
	res, ok := urlCache.Get(conf.Next)
	
	if ok {
		body = res
	} else {
		resp, err := http.Get(conf.Next)
		if err != nil {
			fmt.Println("Error sending Get Request to Location-Area Endpoint: " + err.Error())
			return nil
		}
		body, err = io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode > 299 {
			fmt.Printf("Response failed with status code: %d\n", resp.StatusCode)
			return nil
		} else {
			urlCache.Add(conf.Next, body)
		}
	}
	locationArea := UnmarshaledLocationAreas{}
	err := json.Unmarshal(body, &locationArea)
	if err != nil {
		fmt.Printf("Error processing json response: %s\n", err.Error())
		return nil
	}
	conf.Next = locationArea.Next
	conf.Prev = locationArea.Previous
	for _, area := range locationArea.Results {
		fmt.Println(area.Name)
	}
	return nil
}

func commandMapb(conf *config) error {
	var body []byte
	if conf.Prev == "" {
		fmt.Println("you're on the first page")
		return nil
	}
	body, ok := conf.Cache.Get(conf.Prev)
	if !ok {
		resp, err := http.Get(conf.Prev)
		if err != nil {
			fmt.Println("Error sending Get Request to Location-Area Endpoint: " + err.Error())
		}
		body, err = io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode > 299 {
			fmt.Printf("Request failed with status code: %d\n", resp.StatusCode)
			return nil
		} else {
			conf.Cache.Add(conf.Prev, body)
		}
	}
	locationArea := UnmarshaledLocationAreas{}
	err := json.Unmarshal(body, &locationArea)
	if err != nil {
		fmt.Printf("Error processing json response: %s\n", err.Error())
		return nil
	}
	conf.Next = locationArea.Next
	conf.Prev = locationArea.Previous
	for _, area := range locationArea.Results {
		fmt.Println(area.Name)
	}
	return nil
}
func commandExplore(conf *config) error {
	baseUrl := "https://pokeapi.co/api/v2/location-area/"
	areaToSearchUrl := baseUrl + conf.ExploreArg
	var body []byte
	res, ok := conf.Cache.Get(areaToSearchUrl)
	if ok {
		body = res
	} else {
		resp, err := http.Get(areaToSearchUrl)
		if err != nil {
			fmt.Println("Error sending Get Request to Location-Area Endpoint: " + err.Error())
			return nil
		}
		body, err = io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode > 299 {
			fmt.Printf("Response failed with status code: %d\n", resp.StatusCode)
			return nil
		} else {
			conf.Cache.Add(areaToSearchUrl, body)
		}
	}
	pokemonEncounters := UnmarshaledPokemonEncounters{}
	err := json.Unmarshal(body, &pokemonEncounters)
	if err != nil {
		fmt.Printf("Error processing json response: %s\n", err.Error())
		return nil
	}
	for i := range pokemonEncounters.PokemonEncounters {
		fmt.Println(pokemonEncounters.PokemonEncounters[i].Pokemon.Name)
	}
	return nil
}
func commandCatch(conf *config) error {
	var body []byte
	var speciesBody []byte
	pokemonToCatch := conf.CatchArg
	cache := conf.Cache

	basePokemonUrl := "https://pokeapi.co/api/v2/pokemon/"
	baseSpeciesUrl := "https://pokeapi.co/api/v2/pokemon-species/"
	
	res, ok := cache.Get(basePokemonUrl + pokemonToCatch)
	resSpec, ok_ := cache.Get(baseSpeciesUrl + pokemonToCatch)
	if ok && ok_ {
		body = res
		speciesBody = resSpec
	} else {
		resp, err := http.Get(basePokemonUrl + pokemonToCatch)
		if err != nil {
			fmt.Printf("Error sending Get Request to Pokemon Endpoint: %s\n", err.Error())
			return nil
		}
		if resp.StatusCode > 299 {
			if resp.StatusCode == 404 {
				fmt.Printf("%s is not a Pokemon - please choose a valid Pokemon to catch\n", pokemonToCatch)
				return nil
			}
			fmt.Printf("Received error as response with status code: %d\n", resp.StatusCode)
			return nil
		}
		body, err = io.ReadAll(resp.Body)
		resp.Body.Close()

		respSpecies, err := http.Get(baseSpeciesUrl + pokemonToCatch)
		if err != nil {
			fmt.Printf("Error sending Get Request to Pokemon Endpoint: %s\n", err.Error())
			return nil
		}
		if resp.StatusCode > 299 {
			fmt.Printf("Received error as response with status code: %d\n", resp.StatusCode)
			return nil
		}
		speciesBody, err = io.ReadAll(respSpecies.Body)
		respSpecies.Body.Close()
		// cache both endpoint responses
		cache.Add(basePokemonUrl + pokemonToCatch, body)
		cache.Add(baseSpeciesUrl + pokemonToCatch, speciesBody)
	}

	pokemonData := UnmarshaledPokemonInfo{}
	pokemonSpecies := UnmarshaledPokemonSpecies{}
	err := json.Unmarshal(body, &pokemonData)
	if err != nil {
		fmt.Printf("Error processing json response: %s\n", err.Error())
		return nil
	}
	err = json.Unmarshal(speciesBody, &pokemonSpecies)
	if err != nil {
		fmt.Printf("Error processing json response: %s\n", err.Error())
		return nil
	}
	pokemonData.BaseHappiness = pokemonSpecies.BaseHappiness
	pokemonData.CaptureRate = pokemonSpecies.CaptureRate
	pokemonData.EntryDescr = pokemonSpecies.FlavorText[0].EntryDescr

	fmt.Printf("Throwing a Pokeball at %s...\n", pokemonToCatch)
	if caughtPokemon(pokemonData.CaptureRate) {
		conf.Pokedex[pokemonToCatch] = pokemonData
		fmt.Printf("%s was caught!\n", pokemonToCatch)
		fmt.Printf("%s's data has been added to the pokedex!\n", pokemonToCatch)
	} else {
		fmt.Printf("%s escaped!\n", pokemonToCatch)
	} 
	return nil
}
func caughtPokemon(catchRate int) bool {
	newSource := rand.NewSource(time.Now().UnixNano())
	rand := rand.New(newSource)
	randNum := rand.Intn(425) - 75
	if catchRate >= randNum {
		return true
	} else {
		return false
	}
}
func commandInspect(conf *config) error {
	pokemonName := conf.InspectArg
	pokemonData, ok := conf.Pokedex[pokemonName]
	if !ok {
		fmt.Println("you have not caught that pokemon")
		return nil
	}
	fmt.Println()
	fmt.Printf("Name: %s\n", pokemonName)
	cleanedEntry := strings.ReplaceAll(pokemonData.EntryDescr, string('\f'), " ")
	cleanedEntry = strings.ReplaceAll(cleanedEntry, string('\n'), " ")
	fmt.Printf("Pokedex Entry: %s\n", cleanedEntry)

	if len(pokemonData.Type) == 1 {
		fmt.Printf("Type:\n")
	} else {
		fmt.Printf("Types:\n")
	}
	for i := range pokemonData.Type {
		fmt.Printf("  - %s\n", pokemonData.Type[i].Type.Name)
	}
	fmt.Printf("Abilities:\n")
	for i := range pokemonData.Abilities {
		fmt.Printf("  -%s\n", pokemonData.Abilities[i].Ability.Name)
	}
	fmt.Printf("Stats:\n")
	for i := range pokemonData.BaseStats {
		fmt.Printf("  -%s: %d\n", pokemonData.BaseStats[i].Stat.Name, pokemonData.BaseStats[i].BaseStat)
	}
	fmt.Printf("Height: %v\n", pokemonData.Height)
	fmt.Printf("Weight: %v\n", pokemonData.Weight)
	return nil
}
func commandPokedex(conf *config) error {
	fmt.Println("Your Pokedex:")
	for pokemonName, _ := range conf.Pokedex {
		fmt.Printf(" - %s\n", pokemonName)
	}
	return nil
}
