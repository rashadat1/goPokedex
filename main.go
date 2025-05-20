package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	for {
		_, err := fmt.Fprint(os.Stdout, "Pokedex >")
		if err != nil {
			log.Fatal("Error starting REPL loop" + err.Error())
		}
		inputReader := bufio.NewScanner(os.Stdin)
		inputReader.Scan()
		if inputReader.Err() != nil {
			log.Println("Error reading input: " + inputReader.Err().Error())
		}
		rawInput := inputReader.Text()
		cleanedInput := cleanInput(rawInput)

		fmt.Printf("Your command was: %s\r\n", cleanedInput[0])
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
