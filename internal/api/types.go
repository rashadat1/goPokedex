package api

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
