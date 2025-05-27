package pokemongenerator

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/rashadat1/goPokedex/internal/api"
	"github.com/rashadat1/goPokedex/internal/statCalculator"
)

// we want to take a pokemon species, map of evs, map of ivs, level

type Pokemon struct {
	Species          string
	Level            int
	CurrHp           int
	Moves			 map[string]Move
	Ability			 string
	Nature           string
	Stats            map[string]BundleStats
}
type BundleStats struct {
	StatValue        int
	EVValue          int
	IVValue          int
	EffortValue      int
}
// some of the fields here can be enums and some need to be structs
type Move struct {
	Name             string
	Power            int
	Accuracy         int
	PP               int
	StatChance       int
	DamageClass      string
	Target           string
	Ailment          Ailment
	StatChange       []StatChange
	Description      string
	AilmentChance    int
}
type StatChange struct {
	Stat             string
	Amount           int
	Target           string // self or opponent
}
type Ailment struct {
	Name             string
	Damage           float32
	StatEffect       []AilmentStatChange
}
type AilmentStatChange struct {
	Stat             string
	Change           float32
}

func GeneratePokemon(species string, level int) (Pokemon, error) {
	// method to generate new instance of pokemon - create wild and npc pokemon
	// nature, evs, ivs are random (evs should be zero for wild pokemon)
	// use species to get base 
	newSource := rand.NewSource(time.Now().UnixNano())
	rand := rand.New(newSource)
	statNames := [6]string{"hp", "attack", "defense", "special-attack", "special-defense", "speed"}
	natures := [25]string{
		"Hardy",
		"Lonely",
		"Brave",
		"Adamant",
		"Naughty",
		"Bold",
		"Docile",
		"Relaxed",
		"Impish",
		"Lax",
		"Timid",
		"Hasty",
		"Serious",
		"Jolly",
		"Naive",
		"Modest",
		"Mild",
		"Quiet",
		"Bashful",
		"Rash",
		"Calm",
		"Gentle",
		"Sassy",
		"Careful",
		"Quirky",
	}
	basePokemonUrl := "https://pokeapi.co/api/v2/pokemon/"
	baseSpeciesUrl := "https://pokeapi.co/api/v2/pokemon-species/"

	respPokemon, err := http.Get(basePokemonUrl + species)
	if err != nil {
		fmt.Printf("Error sending Get Request to Pokemon Endpoint: %s\n", err.Error())
		return Pokemon{}, err
	}
	if respPokemon.StatusCode > 299 {
		if respPokemon.StatusCode == 404 {
			fmt.Printf("%s is not a Pokemon - please choose a valid Pokemon\n", species)
			return Pokemon{}, err
		}
		fmt.Printf("Received error as response with status code: %d\n", respPokemon.StatusCode)
		return Pokemon{}, err
	}
	bodyPokemon, err := io.ReadAll(respPokemon.Body)
	respPokemon.Body.Close()

	respSpecies, err := http.Get(baseSpeciesUrl + species)
	if err != nil {
		fmt.Printf("Error sending Get Request to Species Endpoint: %s\n", err.Error())
		return Pokemon{}, err
	}
	if respSpecies.StatusCode > 299 {
		fmt.Printf("Recieved error as response with status code: %d\n", respSpecies.StatusCode)
		return Pokemon{}, err
	}
	bodySpecies, err := io.ReadAll(respSpecies.Body)
	respSpecies.Body.Close()

	pokemonData := api.UnmarshaledPokemonInfo{}
	speciesData := api.UnmarshaledPokemonSpecies{}
	err = json.Unmarshal(bodyPokemon, &pokemonData)
	if err != nil {
		fmt.Printf("Error processing json response from Pokemon Endpoint: %s\n", err.Error())
		return Pokemon{}, err
	}
	err = json.Unmarshal(bodySpecies, &speciesData)
	if err != nil {
		fmt.Printf("Error processing json response from Species Endpoint: %s\n", err.Error())
		return Pokemon{}, err
	}
	pokemonData.BaseHappiness = speciesData.BaseHappiness
	pokemonData.CaptureRate = speciesData.CaptureRate
	pokemonData.EntryDescr = speciesData.FlavorText[0].EntryDescr

	stats := make(map[string]BundleStats)
	for _, stat := range statNames {
		var effort int
		for _, statEntry := range pokemonData.BaseStats {
			if strings.ToLower(statEntry.Stat.Name) == stat {
				effort = statEntry.Effort
			}
		}
		stats[stat] = BundleStats{
			EVValue: 0,
			IVValue: rand.Intn(32),
			EffortValue: effort,
		}
	}
	for _, stat := range statNames {
		if stat == "hp" {
			hpStat := stats["hp"]
			hpStat.StatValue = statCalculator.CalculateHp(pokemonData.BaseStats[0].BaseStat, stats["hp"].IVValue, stats["hp"].EVValue, level)
			stats["hp"] = hpStat
		} else {
			var baseStatVal int
			for _, statEntry := range pokemonData.BaseStats {
				if strings.ToLower(statEntry.Stat.Name) == stat {
					baseStatVal = statEntry.BaseStat
				}
			}
			otherStat := stats[stat]
			otherStat.StatValue = statCalculator.CalculateOtherStat(baseStatVal, stats[stat].IVValue, stats[stat].EVValue, level, 1)
			// 1 is neutral nature modifier
			stats[stat] = otherStat
		}
	}
	numAbilities := len(pokemonData.Abilities)
	

	hpStat := stats["hp"]
	pokemonInstance := Pokemon{	
		Species: species,
		CurrHp: hpStat.StatValue,
		Level: level,
		Stats: stats,
		Nature: natures[rand.Intn(len(natures))],
		Ability: pokemonData.Abilities[rand.Intn(numAbilities)].Ability.Name,
	}

	return pokemonInstance, nil
}

