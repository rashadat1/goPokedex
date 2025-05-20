package main

import "testing"

func TestCleanInput(t *testing.T) {
	cleanInputCases := []struct {
		input       string
		expected    []string
	}{
		{
			input: "  hello  world  ",
			expected: []string{"hello", "world"},
		},
		{
			input: "   charmander,  bulbasaur, SQUIRTLE, ChArMelEoN, Charizard       ",
			expected: []string{"charmander", "bulbasaur", "squirtle", "charmeleon", "charizard"},
		},
		{
			input: "      charmander cHarMELEON CHARIZARD bulbasaur ",
			expected: []string{"charmander", "charmeleon", "charizard", "bulbasaur"},
		},

	}	
	for _, c := range cleanInputCases {
		actual := cleanInput(c.input)
		for i := range actual {
			word := actual[i]
			expectedWord := c.expected[i]
			if word != expectedWord {
				t.Errorf("Got %s expected %s", word, expectedWord)
			}
		}
	}
}

