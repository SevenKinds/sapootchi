package simulation

import "time"

// Direct care actions available from the Home scene. Bathing happens through
// the Scrub mini-game; Rest tucks the pet in; petting is an affection moment.
const (
	batheRestore = 40.0

	pettingBoost    = 4.0
	pettingCooldown = 3 * time.Minute
)

// Bathe raises Hygiene (kept as the plain-sim primitive; the UI reaches it
// through the Scrub mini-game).
func (p *Pet) Bathe() {
	if !p.Awake() {
		return
	}
	p.Stats.Hygiene = clamp(p.Stats.Hygiene + batheRestore)
}

// Rest tucks the pet in: a VOLUNTARY nap. It sleeps (fast energy regen) until
// energy is FULL — unlike the forced nap, which wakes at the threshold. The
// energy pill still wakes it early.
func (p *Pet) Rest() {
	if !p.Awake() {
		return
	}
	p.Asleep = true
	p.NapVoluntary = true
}

// Pet (petting) is an affection moment: the happiness bonus lands at most once
// per cooldown window. Returns whether the bonus was granted — the UI shows
// hearts either way, but spamming pets is love, not progress.
func (p *Pet) Pet(now time.Time) bool {
	if !p.Awake() {
		return false
	}
	if now.Sub(p.LastPetted) < pettingCooldown {
		return false
	}
	p.LastPetted = now
	p.Stats.Happiness = clamp(p.Stats.Happiness + pettingBoost)
	return true
}

// PerfectCare reports whether all four visible stats are full. The Home scene
// awards the once-a-day Perfect Care bonus off this.
func (p *Pet) PerfectCare() bool {
	return p.Stats.Happiness >= 100 &&
		p.Stats.Hunger >= 100 &&
		p.Stats.Hygiene >= 100 &&
		p.Stats.Energy >= 100
}
