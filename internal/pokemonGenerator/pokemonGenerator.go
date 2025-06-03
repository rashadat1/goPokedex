package pokemongenerator

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/rashadat1/goPokedex/internal/api"
	"github.com/rashadat1/goPokedex/internal/statCalculator"
)

// we want to take a pokemon species, map of evs, map of ivs, level



var moveCache = make(map[string]*api.MoveDetail)

func GeneratePokemon(species string, level int) (api.Pokemon, error) {
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
		return api.Pokemon{}, err
	}
	if respPokemon.StatusCode > 299 {
		if respPokemon.StatusCode == 404 {
			fmt.Printf("%s is not a Pokemon - please choose a valid Pokemon\n", species)
			return api.Pokemon{}, err
		}
		fmt.Printf("Received error as response with status code: %d\n", respPokemon.StatusCode)
		return api.Pokemon{}, err
	}
	bodyPokemon, err := io.ReadAll(respPokemon.Body)
	respPokemon.Body.Close()

	respSpecies, err := http.Get(baseSpeciesUrl + species)
	if err != nil {
		fmt.Printf("Error sending Get Request to Species Endpoint: %s\n", err.Error())
		return api.Pokemon{}, err
	}
	if respSpecies.StatusCode > 299 {
		fmt.Printf("Recieved error as response with status code: %d\n", respSpecies.StatusCode)
		return api.Pokemon{}, err
	}
	bodySpecies, err := io.ReadAll(respSpecies.Body)
	respSpecies.Body.Close()

	pokemonData := api.UnmarshaledPokemonInfo{}
	speciesData := api.UnmarshaledPokemonSpecies{}
	err = json.Unmarshal(bodyPokemon, &pokemonData)
	if err != nil {
		fmt.Printf("Error processing json response from Pokemon Endpoint: %s\n", err.Error())
		return api.Pokemon{}, err
	}
	err = json.Unmarshal(bodySpecies, &speciesData)
	if err != nil {
		fmt.Printf("Error processing json response from Species Endpoint: %s\n", err.Error())
		return api.Pokemon{}, err
	}
	pokemonData.BaseHappiness = speciesData.BaseHappiness
	pokemonData.CaptureRate = speciesData.CaptureRate
	pokemonData.EntryDescr = speciesData.FlavorText[0].EntryDescr

	stats := make(map[string]api.BundleStats)
	for _, stat := range statNames {
		var effort int
		for _, statEntry := range pokemonData.BaseStats {
			if strings.ToLower(statEntry.Stat.Name) == stat {
				effort = statEntry.Effort
			}
		}
		stats[stat] = api.BundleStats{
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
	
	typeNames := make([]string, len(pokemonData.Type))
	for i, t := range pokemonData.Type {
		typeNames[i] = t.Type.Name
	}

	hpStat := stats["hp"]
	pokemonInstance := api.Pokemon{	
		Species: species,
		CurrHp: hpStat.StatValue,
		Level: level,
		Type: typeNames,
		Stats: stats,
		Nature: natures[rand.Intn(len(natures))],
		Ability: pokemonData.Abilities[rand.Intn(numAbilities)].Ability.Name,
	}
	moveList := CreateLearnset(species, pokemonData)

	knowableMoves := []string{}

	knowableMoves = append(knowableMoves, moveList.EggMoves...)
	knowableMoves = append(knowableMoves, moveList.MachineMoves...)
	knowableMoves = append(knowableMoves, moveList.TutorMoves...)
	selectedIndices := []int{}
	for i := 0; i <= level; i++ {
		learnedAtLevel, ok := moveList.LevelUpMoves[i]
		if ok {
			knowableMoves = append(knowableMoves, learnedAtLevel...)
		}
	}
	for {
		if len(selectedIndices) == 4 {
			break
		}
		index := rand.Intn(len(knowableMoves))
		if !slices.Contains(selectedIndices, index) {
			selectedIndices = append(selectedIndices, index)
		}

	}
	chosenMoveNames := []string{}
	for i,_ := range selectedIndices {
		chosenMoveNames = append(chosenMoveNames, knowableMoves[selectedIndices[i]])

	}
	chosenMoveInstances := [4]*api.MoveInstance{}
	for i, moveName := range chosenMoveNames {
		moveDetailData := GetMoveDetail(moveName)
		moveInstance := api.MoveInstance{
			RemainingPP: moveDetailData.PP,
			Detail: moveDetailData,
		}
		chosenMoveInstances[i] = &moveInstance
	}
	pokemonInstance.Moves = chosenMoveInstances

	return pokemonInstance, nil
}
func CreateLearnset(species string, pokemonData api.UnmarshaledPokemonInfo) api.MoveList{
	var versionGroups = [25]string{
		"scarlet-violet",
		"legends-arceus",
		"brilliant-diamond-and-shining-pearl",
		"the-crown-tundra",
		"the-isle-of-armor",
		"sword-shield",
		"lets-go-pikachu-lets-go-eevee",
		"ultra-sun-ultra-moon",
		"sun-moon",
		"omega-ruby-alpha-sapphire",
		"x-y",
		"black-2-white-2",
		"xd",
		"colosseum",
		"black-white",
		"heartgold-soulsilver",
		"platinum",
		"diamond-pearl",
		"firered-leafgreen",
		"emerald",
		"ruby-sapphire",
		"crystal",
		"gold-silver",
		"yellow",
		"red-blue",
	}
	moveData := pokemonData.Moves
	
	mostRecentVersion := getMostRecentVersion(moveData, versionGroups)
	
	moveList := api.MoveList{
		LevelUpMoves: make(map[int][]string),
		EggMoves: []string{},
		TutorMoves: []string{},
		MachineMoves: []string{},
	}
	for _, move := range moveData {
		versionDetailsForMove := move.VersionDetails
		for _, versionDetail := range versionDetailsForMove {
			if versionDetail.VersionGroup.Name == mostRecentVersion {
				if versionDetail.MoveLearnMethod.Name == "level-up" {
					_, ok := moveList.LevelUpMoves[versionDetail.LevelLearnedAt]
					if !ok {
						moveList.LevelUpMoves[versionDetail.LevelLearnedAt] = []string{move.Move.Name}
					} else {
						levelUpMoves := moveList.LevelUpMoves[versionDetail.LevelLearnedAt]
						moveList.LevelUpMoves[versionDetail.LevelLearnedAt] = append(levelUpMoves, move.Move.Name)
					}
				} else if versionDetail.MoveLearnMethod.Name == "egg" {
					eggMoves := moveList.EggMoves
					eggMoves = append(eggMoves, move.Move.Name)
					moveList.EggMoves = eggMoves
				} else if versionDetail.MoveLearnMethod.Name == "tutor" {
					tutorMoves := moveList.TutorMoves
					tutorMoves = append(tutorMoves, move.Move.Name)
					moveList.TutorMoves = tutorMoves
				} else if versionDetail.MoveLearnMethod.Name == "machine" {
					machineMoves := moveList.MachineMoves
					machineMoves = append(machineMoves, move.Move.Name)
					moveList.MachineMoves = machineMoves
				}
			}
		}
	}
	return moveList

}
func getMostRecentVersion(moveData []api.MoveData, versionGroups [25]string) string {
	for _, versionName := range versionGroups {
		// we loop through beginning with the most recent version 
		for _, move := range moveData {
			versionDetailsForMove := move.VersionDetails
			for _, versionDetail := range versionDetailsForMove {
				if versionDetail.VersionGroup.Name == versionName {
					return versionName
				}
			}
		}
	}
	return ""
}
func GetMoveDetail(moveName string) *api.MoveDetail {
	moveBaseUrl := "https://pokeapi.co/api/v2/move/"
	fullMoveUrl := moveBaseUrl + moveName
	
	if val, ok := moveCache[moveName]; ok {
		return val
	}

	resp, err := http.Get(fullMoveUrl)
	if err != nil {
		fmt.Println("Error sending Get Request to Move endpoint: " + err.Error())
		return &api.MoveDetail{}
	}
	if resp.StatusCode > 299 {
		if resp.StatusCode == 404 {
			fmt.Printf("%s is not a Pokemon move - please use a valid move name\n", moveName)
		}
		fmt.Printf("Get Request to Move endpoint returned error with status code: %d\n", resp.StatusCode)
		return &api.MoveDetail{}

	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	moveDetailData := api.MoveDetail{}
	err = json.Unmarshal(body, &moveDetailData)
	if err != nil {
		fmt.Println("Error processing json response to Move endpoint: " + err.Error())
		return &api.MoveDetail{}
	}
	moveCache[moveName] = &moveDetailData
	return &moveDetailData
}
