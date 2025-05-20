package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)


var commandRegistry map[string]cliCommand

type cliCommand struct {
	name           string
	description    string
	callback       func() error
}

func main() {
	inputReader := bufio.NewScanner(os.Stdin)
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
		
			commandData, exists := commandRegistry[commandName]
			if exists {
				err = commandData.callback()
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

func commandExit() error {
	// callback for exit command
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil	
}

func commandHelp() error {
	fmt.Print("Welcome to the Pokedex!\r\nUsage:\r\n\r\n")
	for _, value := range commandRegistry {
		fmt.Println(value.name + ": " + value.description)
	}
	return nil
}
