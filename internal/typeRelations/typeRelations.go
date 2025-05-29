package typeRelations

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/rashadat1/goPokedex/internal/api"
)

func GetTypeRelations() (*api.TypeEffect, error){
	// api.TypeEffect is a map of maps where the keys are types and the values are mapps whose
	// keys are types and values are damage multiplier
	baseTypeUrl := "https://pokeapi.co/api/v2/type/"
	typeEffectRes := api.TypeEffect{
		TypeMap: make(map[string]api.Relations),
	}

	for i := 1; i < 19; i++ {

		typeInteractions := make(map[string]float32)
		resp, err := http.Get(baseTypeUrl + strconv.Itoa(i) + "/")
		if err != nil {
			fmt.Println("Error sending Get Request to Types Endpoint: " + err.Error())
			return &api.TypeEffect{}, nil
		}
		if resp.StatusCode > 299 {
			fmt.Println("Received error from Get Request to Type Endpoint: " + strconv.Itoa(resp.StatusCode))
			return &api.TypeEffect{}, nil
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading json body into byte slice: " + err.Error())
			return &api.TypeEffect{}, nil
		}
		resp.Body.Close()
		typeRelationsUnmarshal := api.TypeRelationsUnmarshal{}
		err = json.Unmarshal(body, &typeRelationsUnmarshal)
		if err != nil {
			fmt.Println("Error parsing json body: " + err.Error())
			return &api.TypeEffect{}, nil
		}
		// because type relations and damage are symmetric -> we just need to track the offensive side
		// then we can check what happens when an attack of type x hits a pokemon of type y
		for _, superEffectiveType := range typeRelationsUnmarshal.DamageRelations.DoubleDmgTo {
			typeInteractions[superEffectiveType.Name] = 2
		}
		for _, resistedType := range typeRelationsUnmarshal.DamageRelations.HalfDmgTo {
			typeInteractions[resistedType.Name] = 0.5
		}
		for _, immuneType := range typeRelationsUnmarshal.DamageRelations.NoDmgTo {
			typeInteractions[immuneType.Name] = 0
		}
		
		extractedRelations := api.Relations{
			Effectiveness: typeInteractions,
		}
		typeEffectRes.TypeMap[typeRelationsUnmarshal.Name] = extractedRelations
		
	}
	return &typeEffectRes, nil

}
