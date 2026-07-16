package simulation

// Direct care actions available from the Home scene. Bathe/Rest are simple
// actions for the POC (Hygiene/Energy also have mini-games later). Petting is
// always available and never gated.
const (
	batheRestore   = 40.0
	restRestore    = 40.0
	pettingBoost   = 5.0
)

// Bathe raises Hygiene.
func (p *Pet) Bathe() {
	if !p.Awake() {
		return
	}
	p.Stats.Hygiene = clamp(p.Stats.Hygiene + batheRestore)
}

// Rest raises Energy (a manual fast-forward on top of passive regen).
func (p *Pet) Rest() {
	if !p.Awake() {
		return
	}
	p.Stats.Energy = clamp(p.Stats.Energy + restRestore)
}

// Pet (petting) gives a small Happiness boost.
func (p *Pet) Pet() {
	if !p.Awake() {
		return
	}
	p.Stats.Happiness = clamp(p.Stats.Happiness + pettingBoost)
}

// PerfectCare reports whether all four visible stats are full. The Home scene
// awards the once-a-day Perfect Care bonus off this.
func (p *Pet) PerfectCare() bool {
	return p.Stats.Happiness >= 100 &&
		p.Stats.Hunger >= 100 &&
		p.Stats.Hygiene >= 100 &&
		p.Stats.Energy >= 100
}
