package simulation

import (
	"math/rand"
	"time"
)

// Decay rates in percentage points per 24h of wall-clock time.
//
// [LOCKED] Hunger = 25%/day (full->empty in 4 days). The others are starting
// values to tune — kept gentle so a once-a-day player stays comfortable.
const (
	hungerDecayPerDay    = 25.0
	happinessDecayPerDay = 20.0
	hygieneDecayPerDay   = 15.0
	// Energy does NOT decay — it regenerates (see energy.go).

	// offlineClampHours caps how much decay a single absence can apply, so
	// returning after a long time is sad-but-recoverable, never hopeless.
	offlineClampHours = 48.0

	// neglectCascadeMultiplier: while any stat sits at 0, Happiness decays faster.
	neglectCascadeMultiplier = 2.0

	// returnChancePerDay is the daily chance an away pet comes back while food
	// is left out. [LOCKED] 28%.
	returnChancePerDay = 0.28
)

// Update advances the pet by real elapsed wall-clock time. Call it every tick;
// it is safe to call at any frequency because it works off LastSeen, not frames.
//
// rng is used only for the away-return roll and may be nil while the pet is home.
func (p *Pet) Update(now time.Time, rng *rand.Rand) {
	if p.Away {
		p.tryReturn(now, rng)
		return
	}
	p.applyDecay(now)
}

func (p *Pet) applyDecay(now time.Time) {
	elapsed := now.Sub(p.LastSeen)
	if elapsed <= 0 {
		p.LastSeen = now
		return
	}
	hours := elapsed.Hours()
	if hours > offlineClampHours {
		hours = offlineClampHours
	}
	days := hours / 24.0

	// Decaying stats. Neglect cascade: a filthy pet (Hygiene 0) loses happiness
	// faster. (Hunger 0 runs him away outright, so it needs no cascade.)
	happinessRate := happinessDecayPerDay
	if p.Stats.Hygiene <= 0 {
		happinessRate *= neglectCascadeMultiplier
	}
	p.Stats.Hunger = clamp(p.Stats.Hunger - hungerDecayPerDay*days)
	p.Stats.Happiness = clamp(p.Stats.Happiness - happinessRate*days)
	p.Stats.Hygiene = clamp(p.Stats.Hygiene - hygieneDecayPerDay*days)

	// Energy is inverted: it regenerates (idle, or fast while asleep) and is
	// spent by activity. Evaluate the sleep trigger BEFORE regen so hitting 0
	// naps him; wake once regen carries him back to the threshold.
	if !p.Asleep && p.Stats.Energy <= 0 {
		p.Asleep = true
	}
	regen := energyIdleRegenPerDay
	if p.Asleep {
		regen = energySleepRegenPerDay
	}
	p.Stats.Energy = clamp(p.Stats.Energy + regen*days)
	if p.Asleep && p.Stats.Energy >= energyWakeThreshold {
		p.Asleep = false
	}

	p.LastSeen = now

	// Soft-fail: hunger empty -> the pet runs away.
	if p.Stats.Hunger <= 0 {
		p.runAway()
	}
}

func (p *Pet) runAway() {
	p.Away = true
	p.FoodLeftOut = false
	// LastSeen already == now (set by applyDecay); it becomes the "left at" mark.
}

// LeaveFoodOut is the player's response to the runaway prompt: leave food out to
// attract him back. Each following day there is a returnChancePerDay chance.
func (p *Pet) LeaveFoodOut() {
	if p.Away {
		p.FoodLeftOut = true
	}
}

func (p *Pet) tryReturn(now time.Time, rng *rand.Rand) {
	if !p.FoodLeftOut || rng == nil {
		return // no rolls until the player leaves food out
	}
	// Roll once per whole elapsed day, advancing LastSeen by a day each roll so
	// the sub-day remainder is preserved across frames.
	for now.Sub(p.LastSeen) >= 24*time.Hour {
		p.LastSeen = p.LastSeen.Add(24 * time.Hour)
		if rng.Float64() < returnChancePerDay {
			p.returnHome(now)
			return
		}
	}
}

func (p *Pet) returnHome(now time.Time) {
	p.Away = false
	p.FoodLeftOut = false
	p.LastSeen = now
	// He came back for the food that was left out: one standard portion.
	p.Stats.Hunger = clamp(p.Stats.Hunger + standardFoodHunger)
}
