package simulation

// Mood is derived from stats (and later, recent events). It affects ONLY
// animation + dialogue — never gameplay.
type Mood int

const (
	MoodHappy Mood = iota
	MoodHungry
	MoodExcited
	MoodBored
	MoodSleepy
	MoodCurious
	MoodLonely
)

func (m Mood) String() string {
	switch m {
	case MoodHappy:
		return "Happy"
	case MoodHungry:
		return "Hungry"
	case MoodExcited:
		return "Excited"
	case MoodBored:
		return "Bored"
	case MoodSleepy:
		return "Sleepy"
	case MoodCurious:
		return "Curious"
	case MoodLonely:
		return "Lonely"
	default:
		return "Unknown"
	}
}

const lowStat = 25.0

// Mood derives current mood from stats using a fixed precedence: the most
// pressing need wins. Excited/Curious are event-driven and layered on elsewhere.
func (p *Pet) Mood() Mood {
	switch {
	case p.Asleep:
		return MoodSleepy
	case p.Stats.Hunger < lowStat:
		return MoodHungry
	case p.Energized():
		return MoodExcited
	case p.Stats.Energy < lowStat:
		return MoodSleepy
	case p.Stats.Happiness < lowStat:
		return MoodLonely
	case p.Stats.Happiness >= 80:
		return MoodHappy
	default:
		return MoodBored
	}
}
