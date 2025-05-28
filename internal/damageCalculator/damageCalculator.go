package damageCalculator

import (
	"math/rand"
	"time"

	pokemongenerator "github.com/rashadat1/goPokedex/internal/pokemonGenerator"
)

func BasicDamageCalculator(attacker, defender pokemongenerator.Pokemon) int {
	power := 50
	inner := ((((2 * attacker.Level) / 5 + 2) * power * (attacker.Stats["attack"].StatValue / defender.Stats["defense"].StatValue)) / 50) + 2
	
	newSource := rand.NewSource(time.Now().UnixNano())
	rand := rand.New(newSource)
	random := rand.Intn(16) + 85
	
	return inner * random / 100
}
// the full damage calculator takes into account type effectiveness, critical hit, burn, STAB, Weather among other
