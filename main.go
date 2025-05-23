package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
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
	cache          *pokecache.Cache
	exploreArg     string
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
type MethodData struct {
	Name           string `json:"name"`
	Url            string `json:"url"`
}
type EncounterData struct {
	Chance         int `json:"chance"`
	MaxLevel       int `json:"max_level"`
	MinLevel       int `json:"min_level"`
	Method         MethodData `json:"method"`
}
type VersionDetails struct {
	EncounterData  []EncounterData `json:"encounter_details"`	
}
type PokemonIdentity struct {
	Name           string `json:"name"`
	Url            string `json:"url"`
}
type EncounterDetails struct {
	Pokemon         PokemonIdentity `json:"pokemon"`
	VersionDetail   []VersionDetails `json:"version_details"`
}
type UnmarshaledPokemonEncounters struct {
	PokemonEncounters      []EncounterDetails `json:"pokemon_encounters"`
}
// add struct tags so json decoder can match the Go field with the JSON field

func main() {
	inputReader := bufio.NewScanner(os.Stdin)
	cache := pokecache.NewCache(5 * time.Second)
	configuration := config{
		Next: "https://pokeapi.co/api/v2/location-area?offset=0&limit=20",
		Prev: "",
		cache: cache,
		exploreArg: "",
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
	for {
		_, err := fmt.Fprint(os.Stdout, "Pokedex >")
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
			if commandName == "explore" {
				if len(cleanedInput) != 2 {
					fmt.Printf("explore command takes 1 argument %d\n were given", len(cleanedInput) - 1)
					continue
				} else {
					configuration.exploreArg = cleanedInput[1]
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
	urlCache := conf.cache
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
	body, ok := conf.cache.Get(conf.Prev)
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
			conf.cache.Add(conf.Prev, body)
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
	areaToSearchUrl := baseUrl + conf.exploreArg

	body, ok := conf.cache.Get(areaToSearchUrl)
	if !ok {
		resp, err := http.Get(areaToSearchUrl)
		
	}
	return nil
}
