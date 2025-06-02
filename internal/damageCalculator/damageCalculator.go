package damageCalculator

import (
	"math/rand"

	"github.com/rashadat1/goPokedex/internal/api"
	pokemongenerator "github.com/rashadat1/goPokedex/internal/pokemonGenerator"
)

type MoveOutcome struct {
	Damage              int
	Missed              bool
	Flinched            bool
	CausedStatus        string // paralysis, burn, sleep, etc this can be null
	StatusDuration      int // only for freeze/sleep
	TargetStatChanges   map[string]int
	UserStatChanges     map[string]int
	RecoilDamage        int
	NumHits             int // double-slap / bullet-seed
	NumTurns            int // rollout / uproar

}
func BasicDamageCalculator(attacker, defender pokemongenerator.Pokemon, battleContext *api.BattleContext) int {
	power := 50
	inner := ((((2 * attacker.Level) / 5 + 2) * power * (attacker.Stats["attack"].StatValue / defender.Stats["defense"].StatValue)) / 50) + 2
	rng := battleContext.Rng	
	random := rng.Intn(16) + 85
	
	return inner * random / 100
}
// the full damage calculator takes into account type effectiveness, critical hit, burn, STAB, Weather among other

func DamageCalculator(attacker, defender pokemongenerator.Pokemon, typeRelations api.TypeEffect, moveIndexChose int) int {
	return 42
}
func HandleMoveCategories(attacker, defender *pokemongenerator.Pokemon, move *api.MoveDetail, battleContext api.BattleContext) *MoveOutcome {
	
	didHit := handleAccuracyCheck(attacker, defender, move, battleContext)
	if !didHit {
		return &MoveOutcome{Missed: true}
	}
	moveCategory := move.Meta.Category.Name
	moveOutcome := MoveOutcome{
		TargetStatChanges: make(map[string]int),
		UserStatChanges: make(map[string]int),
		Missed: false,
	}
	switch moveCategory {
	case "damage":
		
		if move.Meta.MinHits != 0 {
			numHits := calcExecutedHits(move.Meta.MinHits, move.Meta.MaxHits, battleContext.Rng)
			moveOutcome.NumHits = numHits
		} else {
			moveOutcome.NumHits = 1
		}
		if move.Meta.FlinchChance != 0 {
			causeFlinch := rollFlinched(move.Meta.FlinchChance, battleContext.Rng)
			moveOutcome.Flinched = causeFlinch
		}
	}
	return &moveOutcome
}

func calcExecutedHits(minHits, maxHits int, rng *rand.Rand) int {
	if minHits == maxHits {
		return minHits
	}
	var numHits int
	randomNum := rng.Intn(101)
	if randomNum < 35 {
		numHits = 2
	} else if randomNum < 70 {
		numHits = 3
	} else if randomNum < 85 {
		numHits = 4
	} else {
		numHits = 5
	}
	return numHits
}

func rollFlinched(flinchChance int, rng *rand.Rand) bool{
	return rng.Intn(100) < flinchChance
}

func handleAccuracyCheck(attacker, defender *pokemongenerator.Pokemon, move *api.MoveDetail, battleContext api.BattleContext) bool {
	if move.Accuracy == 0 {
		// exempt from accuracy check - target the user or never miss (ie aerial ace // shockwave)
		return true
	}
	modAccuracy := min(float64(move.Accuracy) * getAccuracyMultiplier(attacker.AccuracyStage) / getAccuracyMultiplier(defender.EvasionStage), 100)
	rng := battleContext.Rng

	return float64(rng.Intn(100)) < modAccuracy

}
func getAccuracyMultiplier(stage int) float64 {
    switch stage {
    case -6:
        return 3.0 / 9
    case -5:
        return 3.0 / 8
    case -4:
        return 3.0 / 7
    case -3:
        return 3.0 / 6
    case -2:
        return 3.0 / 5
    case -1:
        return 3.0 / 4
    case 0:
        return 1.0
    case 1:
        return 4.0 / 3
    case 2:
        return 5.0 / 3
    case 3:
        return 6.0 / 3
    case 4:
        return 7.0 / 3
    case 5:
        return 8.0 / 3
    case 6:
        return 9.0 / 3
    default:
        return 1.0
    }
}
