package damageCalculator

import (
	"math/rand"
	"slices"

	"github.com/rashadat1/goPokedex/internal/api"
)

type MoveOutcome struct {
	Damage                    int
	StatusDuration            int // only for freeze/sleep
	NumHits                   int // double-slap / bullet-seed
	NumTurns                  int // rollout / uproar
	RecoilDamageMultiplier    float32
	CausedStatus              string // paralysis, burn, sleep, etc this can be null
	TargetStatChanges         map[string]int
	UserStatChanges           map[string]int
	Missed                    bool
	Flinched                  bool
	Charging                  bool
}
// special handling for these classes of moves
var RampageMoves = map[string]bool{
    "outrage": true,
    "thrash": true,
    "petal-dance": true,
	"raging-fury": true,
}
// semi-invulnerable moves are a subcategory of this list
var ChargingMoves = map[string]bool{
    "solar-beam": true,
    "skull-bash": true,
    "sky-attack": true,
	"meteor-beam": true,
	"razor-wind": true,
	"bounce": true,
	"dig": true,
	"dive": true,
	"phantom-force": true,
	"electro-shot": true,
	"fly": true,
	"shadow-force": true,
	"freeze-shock": true,
	"sky-drop": true,
	"solar-blade": true,
	"geomancy": true,
	"ice-burn": true,
}
var RechargingMoves = map[string]bool{
    "hyper-beam": true,
    "giga-impact": true,
	"blast-burn": true,
	"hydro-cannon": true,
	"frenzy-plant": true,
	"rock-wrecker": true,
	"roar-of-time": true,
	"meteor-assault": true,
}
var LockInMoves = map[string]bool{
    "rollout": true,
    "ice-ball": true,
}
var TrappingMoves = map[string]bool{
    "wrap": true,
    "bind": true,
    "clamp": true,
    "fire-spin": true,
    "whirlpool": true,
    "sand-tomb": true,
    "magma-storm": true,
	"infestation": true,
}
var ConditionalPowerDoubling = map[string]bool{
	"brine": true,
	"venoshock": true,
	"hex": true,
}
var HighDamageHpLevel = map[string]bool{
	"eruption": true,
	"water-spout": true,
	"dragon-energy": true,
	"flail": true,
	"reversal": true,
}
var PowerBasedOnSpeed = map[string]bool{
	"gyro-ball": true,
	"electro-ball": true,
}
var PowerBasedOnWeightDiff = map[string]bool{
	"heavy-slam": true,
	"heat-crash": true,
}
var MoreDamageHeavy = map[string]bool{
	"low-kick": true,
	"grass-knot": true,
}
var CrashDamageIfMiss = map[string]bool{
	"jump-kick": true,
	"supercell-slam": true,
	"high-jump-kick": true,
}
var MovesWithSemiInvulnerability = map[string]bool{
	"fly": true,
	"bounce": true,
	"sky-drop": true,
	"dig": true,
	"dive": true,
	"shadow-force": true,
	"phantom-force": true,
}
var MovesDamagingSemiVulnerable = map[string][]string{
	"fly": {"gust", "hurricane", "sky-uppercut", "smack-down", "thousand-arrows", "thunder", "twister"},
	"bounce": {"gust", "hurricane", "sky-uppercut", "smack-down", "thousand-arrows", "thunder", "twister"},
	"sky-drop": {"gust", "hurricane", "sky-uppercut", "smack-down", "thousand-arrows", "thunder", "twister"},
	"dig": {"earthquake", "magnitude", "fissure"},
	"dive": {"surf", "whirlpool"},
}
var AlwaysHitsUsedByPoisonType = map[string]bool{
	"toxic": true,
}


func BasicDamageCalculator(attacker, defender api.Pokemon, battleContext *api.BattleContext) int {
	power := 50
	inner := ((((2 * attacker.Level) / 5 + 2) * power * (attacker.Stats["attack"].StatValue / defender.Stats["defense"].StatValue)) / 50) + 2
	rng := battleContext.Rng	
	random := rng.Intn(16) + 85
	
	return inner * random / 100
}
// the full damage calculator takes into account type effectiveness, critical hit, burn, STAB, Weather among other

func DamageCalculator(attacker, defender api.Pokemon, typeRelations api.TypeEffect, moveIndexChose int) int {
	return 42
}
func HandleMoveCategories(attacker, defender *api.Pokemon, move *api.MoveDetail, battleContext api.BattleContext) *MoveOutcome {
	battleState := battleContext.PokemonStates[attacker]
	if ChargingMoves[move.Name] {
		if MovesWithSemiInvulnerability[move.Name] {
			battleState.SemiInvuln = &api.SemiInvulnState{
				Move: move,
				Turn: 1,
			}
		}
		battleState.Charging = &api.ChargingState{
			Move: move,
			NumTurns: 2,
			CurrentTurns: 1,
		}
		battleState.ActiveMove = move.Name
		battleContext.PokemonStates[attacker] = battleState
		return &MoveOutcome{Charging: true}
	}
	didHit := handleAccuracyCheck(attacker, defender, move, battleContext)
	if !didHit {
		return &MoveOutcome{Missed: true}
	}
	moveOutcome := MoveOutcome{
		TargetStatChanges: make(map[string]int),
		UserStatChanges: make(map[string]int),
		Missed: false,
	}
	
	if RampageMoves[move.Name] {
		battleState.Rampaging = &api.RampageState{
			Move: move,
			MaxTurns: battleContext.Rng.Intn(2) + 2, // random number either 2 or 3 
			CurrentTurns: 1,
			WillConfuse: true,

		}
		battleState.ActiveMove = move.Name
		battleContext.PokemonStates[attacker] = battleState
	}
	
	handleMultiHit(move.Meta.MinHits, move.Meta.MaxHits, battleContext.Rng, &moveOutcome)
	handleFlinch(move.Meta.FlinchChance, battleContext.Rng, &moveOutcome)
	handleRecoil(move.Meta.Drain, &moveOutcome)

	return &moveOutcome
}

func handleMultiHit(minHits, maxHits int, rng *rand.Rand, moveOutcome *MoveOutcome) {
	if minHits == 0 && maxHits == 0 {
		moveOutcome.NumHits = 1
		return
	}
	if minHits == maxHits {
		moveOutcome.NumHits = minHits
		return
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
	moveOutcome.NumHits = numHits
	return
}

func handleFlinch(flinchChance int, rng *rand.Rand, moveOutcome *MoveOutcome) {
	causedFlinch := rng.Intn(100) < flinchChance
	moveOutcome.Flinched = causedFlinch
}

func handleRecoil(recoilPercent int, moveOutcome *MoveOutcome) {
	moveOutcome.RecoilDamageMultiplier = -float32(recoilPercent) / 100
}
func handleAccuracyCheck(attacker, defender *api.Pokemon, move *api.MoveDetail, battleContext api.BattleContext) bool {
	if slices.Contains(attacker.Type, "poison") && move.Name == "toxic" {
		// toxic always hits when used by a poison type
		return true
	}

	if semiInvulnData := battleContext.PokemonStates[defender].SemiInvuln; semiInvulnData != nil {
		semiInvulnMove := semiInvulnData.Move.Name
		if slices.Contains(MovesDamagingSemiVulnerable[semiInvulnMove], move.Name) {
			return true
		}
		return false
	}

	if move.Accuracy == 0 {
		return true // moves exempt from normal accuracy calculation e.g. swift and aerial ace
	}

	accuracy := min(float64(move.Accuracy) * getAccuracyMultiplier(attacker.AccuracyStage) / getAccuracyMultiplier(defender.EvasionStage), 100)
	rng := battleContext.Rng

	return float64(rng.Intn(100)) < accuracy

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
