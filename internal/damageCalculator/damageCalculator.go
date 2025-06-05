package damageCalculator

import (
	"fmt"
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
	Message                   string 
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
	"focus-punch": true,
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
var CrushGrip = map[string]bool{
	"crush-grip": true,
	"hard-press": true,
	"wring-out": true,
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
var FixedDamage = map[string]bool{
	"dragon-rage": true,
	"sonic-boom": true,
}
var CounterAttacks = map[string]bool{
	"bide": true,
	"comeuppance": true,
	"counter": true,
	"metal-burst": true,
	"mirror-coat": true,
}
var DirectDamageAttacks = map[string]bool{
	
}
var MovesStrongerAgainstMinimized = map[string]bool{
	"astonish": true,
	"body-slam": true,
	"double-iron-bash": true,
	"dragon-rush": true,
	"extrasensory": true,
	"flying-press": true,
	"heat-crash": true,
	"heavy-slam": true,
	"malicious-moonsault": true,
	"needle-arm": true,
	"phantom-force": true,
	"shadow-force": true,
	"steamroller": true,
	"stomp": true,
}
var MovesCallingOtherMoves = map[string]bool{
	"assist": true,
	"copycat": true,
	"me-first": true,
	"metronome": true,
	"mirror-move": true,
	"nature-power": true,
	"sleep-talk": true,
	"snatch": true,
}
var MovesDealingDamageBasedOnIncreasedStatStages = map[string]bool{
	"punishment": true,
	"power-trip": true,
}
var MovesThatCanHealNonVolatileStatus = map[string]bool{
	"aromatherapy": true,
	"heal-bell": true,
	"jungle-healing": true,
	"lunar-blessing": true,
	"psycho-shift": true,
	"purify": true,
	"refresh": true,
	"rest": true,
	"smelling-salts": true,
	"sparkling-aura": true,
	"sparkly-swirl": true,
	"take-heart": true,
	"uproar": true,
	"wake-up-slap": true,
}
var MovesThatRemoveTypeImmunities = map[string]bool{
	"foresight": true,
	"gravity": true,
	"miracle-eye": true,
	"odor-sleuth": true,
	"smack-down": true,
	"thousand-arrows": true,
}
var MoveAffectedByFriendship = map[string]bool{
	"frustration": true,
	"pika-papow": true,
	"return": true,
	"veevee-volley": true,
}
var ScreenCreatingMoves = map[string]bool{
	"aurora-veil": true,
	"baddy-bad": true,
	"glitzy-glow": true,
	"light-screen": true,
	"reflect": true,
}
var ScreenRemovingMoves = map[string]bool{
	"brick-break": true,
	"defog": true,
	"psychic-fangs": true,
	"raging-bull": true,
}
var MovesThatSwitchTheTargetOut = map[string]bool{
	"circle-throw": true,
	"dragon-tail": true,
	"roar": true,
	"whirlwind": true,
}
var WeatherChangingMoves = map[string]bool{
	"chilly-reception": true,
	"defog": true,
	"hail": true,
	"rain-dance": true,
	"sunny-day": true,
	"sandstorm": true,
	"snowscape": true,
}
var MovesThatSwitchTheUserOut = map[string]bool{
	"baton-pass": true,
	"chilly-reception": true,
	"flip-turn": true,
	"parting-shot": true,
	"shed-tail": true,
	"teleport": true,
	"u-turn": true,
	"volt-switch": true,
}
var EntryHazardRemove = map[string]bool{
	"defog": true,
	"mortal-spin": true,
	"rapid-spin": true,
	"tidy-up": true,
}
var EntryHazardMoves = map[string]bool{
	"spikes": true,
	"stealth-rock": true,
	"sticky-web": true,
	"stone-axe": true,
	"toxic-spikes": true,
	"ceaseless-edge": true,
}
var NonStandardStatUsage = map[string]bool{
	"body-press": true,
	"foul-play": true,
	"psyshock": true,
	"psystrike": true,
	"secret-sword": true,
}
var MovesThatRaiseCritRate = map[string]bool{
	"dragon-cheer": true,
	"focus-energy": true,
	"triple-arrows": true,
}
var MovesThatBreakProtection = map[string]bool{
	"feint": true,
	"hyperspace-fury": true,
	"hyperspace-hole": true,
	"phantom-force": true,
	"shadow-force": true,
}
var MovesThatDamageThroughProtection = map[string]bool{
	"hyper-drill": true,
	"mighty-cleave": true,
}
var MovesWithSpecialTypeEffectiveness = map[string]bool{
	"flying-press": true,
	"freeze-dry": true,
	"thousand-arrows": true,
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
func InitializedSpecialMove(attacker, defender *api.Pokemon, moveInst *api.MoveInstance, battleContext api.BattleContext) int {
	attackerState := battleContext.PokemonStates[attacker]
	defenderState := battleContext.PokemonStates[defender]

	move := moveInst.Detail

	if ChargingMoves[move.Name] {
		if MovesWithSemiInvulnerability[move.Name] {
			attackerState.SemiInvuln = &api.SemiInvulnState{
				Move: move,
				Turn: 1,
			}
			attackerState.ActiveMoveKind = "SemiInvuln"
		} else {
			attackerState.Charging = &api.ChargingState{
				Move: move,
				NumTurns: 2,
				CurrentTurns: 1,
			}
		}

		attackerState.ActiveMove = move.Name
		attackerState.ActiveMoveKind = "Charging"
		
		battleContext.PokemonStates[attacker] = attackerState
		return 1
	}
	didHit := handleAccuracyCheck(attacker, defender, move, battleContext)
	if !didHit {
		return 2
	}
	if RampageMoves[move.Name] {
		attackerState.Rampaging = &api.RampageState{
			Move: move,
			MaxTurns: battleContext.Rng.Intn(2) + 2, // random number either 2 or 3 
			CurrentTurns: 1,
			WillConfuse: true,

		}
		attackerState.ActiveMove = move.Name
		attackerState.ActiveMoveKind = "Rampage"
		battleContext.PokemonStates[attacker] = attackerState
		return 1
	}
	if TrappingMoves[move.Name] {
		defenderState.Trapped = &api.TrappedState{
			Move: move,
			CurrentTurns: 1,
			MaxTurns: battleContext.Rng.Intn(2) + 4,
		}
		defenderState.CanFlee = false
		battleContext.PokemonStates[defender] = defenderState
		return 1
	}
	if LockInMoves[move.Name] {
		attackerState.LockedIn = &api.LockedInState{
			Move: move,
			MaxTurns: 5,
			CurrentTurns: 1,
		}
		attackerState.ActiveMove = move.Name
		attackerState.ActiveMoveKind = "LockedIn"
		battleContext.PokemonStates[attacker] = attackerState
		return 1
	}
	return 0

}
func HandleMoveExecution(attacker, defender *api.Pokemon, moveInst *api.MoveInstance, battleContext *api.BattleContext) *MoveOutcome {
	moveInst.RemainingPP--
	attackerState := battleContext.PokemonStates[attacker]
	move := moveInst.Detail
	moveOutcome := &MoveOutcome{
		TargetStatChanges: make(map[string]int),
		UserStatChanges: make(map[string]int),
	}

	switch attackerState.ActiveMoveKind {
	// handle the case of a semi-invulnerable or charging move on its turn one no miss is possible
	case "SemiInvuln":
		semiInvulnData := attackerState.SemiInvuln
		if semiInvulnData.Turn == 1 {
			printSemiInvulnMessage(attacker.Species, defender.Species, move.Name, moveOutcome)
			semiInvulnData.Turn++
			return moveOutcome
		}
	case "Charging":
		chargingData := attackerState.Charging
		if chargingData.CurrentTurns < chargingData.NumTurns {
			printChargingMessage(attacker.Species, move.Name, move.Name)
			chargingData.CurrentTurns++
			return moveOutcome
		}
	}
	didHit := handleAccuracyCheck(attacker, defender, move, *battleContext)
	if !didHit {
		return &MoveOutcome{
			Missed: true,
		}
	}
	// if the move did not miss then all moves are handled as they should be
	moveOutcome.Missed = false
	
	handleMultiHit(move.Meta.MinHits, move.Meta.MaxHits, battleContext.Rng, moveOutcome)
	handleFlinch(move.Meta.FlinchChance, battleContext.Rng, moveOutcome)	
	
	damageEngine(attacker, defender, moveInst, battleContext, moveOutcome)
	handleRecoil(move.Meta.Drain, moveOutcome)
	
	return moveOutcome
}

func printSemiInvulnMessage(attackerName, defenderName, moveName string) {
	switch moveName {
	case "fly":
		fmt.Printf("%s flew up high!\n", attackerName)
	case "bounce":
		fmt.Printf("%s sprang up!\n", attackerName)
	case "sky-drop":
		fmt.Printf("%s took the enemy %s into the sky!", attackerName, defenderName)
	case "dig":
		fmt.Printf("%s burrowed its way under the ground!", attackerName)
	case "dive":
		fmt.Printf("%s hid underwater!", attackerName)
	}
}

func damageEngine(attacker, defender *api.Pokemon, moveInst *api.MoveInstance, battleContext *api.BattleContext, moveOutcome *MoveOutcome) {
	moveData := moveInst.Detail
	moveCategory := moveData.Meta.Category.Name
	switch moveCategory {
	default:
		calcDamage(attacker, defender, moveInst, battleContext, moveOutcome)
		mutateState()
		calcStatBoost()
		calcHeal()
		calcAilment()
	}

}

func calcDamage(attacker, defender *api.Pokemon, moveInst *api.MoveInstance, battleContext *api.BattleContext, moveOutcome *MoveOutcome) {
	getMovePower(attacker, defender, moveInst, battleContext)

}
func getMovePower(attacker, defender *api.Pokemon, moveInst *api.MoveInstance, battleContext *api.BattleContext) int{
	apiPower := moveInst.Detail.Power
	if apiPower == 0 {

	}
}
func mutateState() {
}
func calcStatBoost() {
}
func calcHeal() {
}
func calcAilment() {
}
func printChargingMessage(attackerName, defenderName, moveName string) {
	switch moveName {
	case "solar-beam":
		fmt.Printf("%s absorbed light!\n", attackerName)
	case "skull-bash":
		fmt.Printf("%s lowered its head!\n", attackerName)
	case "sky-attack":
		fmt.Printf("%s became cloaked in a harsh light!\n", attackerName)
	case "meteor-beam":
		fmt.Printf("%s is overflowing with space power!\n", attackerName)
	case "razor-wind":
		fmt.Printf("%s made a whirlwind!\n", attackerName)
	case "bounce":
		fmt.Printf("%s sprang up!\n", attackerName)
	case "dig":
		fmt.Printf("%s dug a hole!\n", attackerName)
	case "dive":
		fmt.Printf("%s hid underwater!\n", attackerName)
	case "phantom-force":
		fmt.Printf("%s vanished instantly!\n", attackerName)
	case "electro-shot":
		fmt.Printf("%s absorbed electricity!\n", attackerName)
	case "fly":
		fmt.Printf("%s flew up high!\n", attackerName)
	case "shadow-force":
		fmt.Printf("%s vanished instantly!\n", attackerName)
	case "freeze-shock":
		fmt.Printf("%s became cloaked in a freezing light!\n", attackerName)
	case "sky-drop":
		fmt.Printf("%s took the enemy %s into the sky!\n", attackerName, defenderName)
	case "solar-blade":
		fmt.Printf("%s absorbed light!\n", attackerName)
	case "geomancy":
		fmt.Printf("%s is absorbing power!\n", attackerName)
	case "ice-burn":
		fmt.Printf("%s became cloaked in freezing air!\n", attackerName)
	case "focus-punch":
		fmt.Printf("%s is tightening its focus!\n", attackerName)
	}
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
