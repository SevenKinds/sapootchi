package simulation

// Energy is INVERTED relative to the other three stats: it regenerates over
// wall-clock time and is spent by activity (mini-games), rather than decaying.
//
//	idle           -> energy rises slowly
//	resting/asleep -> energy rises fast
//	mini-games     -> energy is spent
//
// Two soft-states sit on the energy axis:
//   - Energized (Energy >= 100): hyper. Idle only refills it, so the ONLY way
//     down is to play a mini-game. A needy "let's play!" moment.
//   - Asleep (triggered at Energy 0): falls asleep on the spot, can't be
//     interacted with, sleeps until Energy reaches energyWakeThreshold (50).
//     Hysteresis: sleep at 0, wake at 50.
const (
	// [TODO] tune these three.
	energyIdleRegenPerDay  = 40.0  // "doing nothing"
	energySleepRegenPerDay = 200.0 // resting / forced sleep
	energyWakeThreshold    = 50.0
)

// Energized reports the full-energy hyper state (must burn it off in a game).
func (p *Pet) Energized() bool { return !p.Asleep && p.Stats.Energy >= 100 }

// Awake reports whether the pet can be interacted with right now.
func (p *Pet) Awake() bool { return !p.Asleep && !p.Away }
