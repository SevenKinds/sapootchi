package simulation

import "time"

// evolveAge is how long a baby stays a baby before it can branch. Placeholder
// value for the POC — tune later.
const evolveAge = 3 * 24 * time.Hour

// [TODO] Full branching tree (Cute->Royal, Smart->Wizard, Wild->Ninja) is v2.
// For the POC we ship ONE evolution step and stub the branch selection: at
// evolveAge, the baby evolves based on which dimension it was cared for most.
//
//	high Intelligence  -> Smart
//	high Energy/runner  -> Wild
//	otherwise (happiness-led) -> Cute

// Age returns the pet's wall-clock age at now.
func (p *Pet) Age(now time.Time) time.Duration {
	return now.Sub(p.BornAt)
}

// CanEvolve reports whether the baby is old enough to take its first step.
func (p *Pet) CanEvolve(now time.Time) bool {
	return p.Phase == PhaseBaby && p.Age(now) >= evolveAge
}

// Evolve advances a baby one step down a stubbed branch. No-op if not eligible.
func (p *Pet) Evolve(now time.Time) {
	if !p.CanEvolve(now) {
		return
	}
	switch {
	case p.Hidden.Intelligence >= 60:
		p.Phase = PhaseSmart
	case p.Stats.Energy >= 80:
		p.Phase = PhaseWild
	default:
		p.Phase = PhaseCute
	}
}
