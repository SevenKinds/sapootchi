package simulation

import "errors"

// standardFoodHunger is the baseline restore: one standard item = +25% hunger,
// so 4 items bring the pet from empty to full. Matches hunger's 25%/day decay:
// one item = one day of hunger.
const standardFoodHunger = 25.0

// FoodKind identifies an inventory item.
type FoodKind int

const (
	FoodApple FoodKind = iota // standard
	FoodSandwich
	FoodCake
	FoodCoffee     // energy consumable, not really "food"
	FoodEnergyPill // dev/test item: full energy refill
)

// FoodDef is the effect of consuming one item. Verb is how the UI phrases
// giving it ("" = Feed). WakesUp items can be given to a SLEEPING pet and end
// the nap — everything else respects it.
type FoodDef struct {
	Name      string
	Verb      string
	Hunger    float64
	Happiness float64
	Energy    float64
	WakesUp   bool
}

// Foods is the item catalog. Apple is the standard (+25%). Coffee restores
// energy (one Runner play's worth plus a little) — note it cannot wake an
// asleep pet; the forced nap stays sacred. The Energy Pill is a DEV item:
// never sold, never won — only spawned from dev mode.
var Foods = map[FoodKind]FoodDef{
	FoodApple:      {Name: "Apple", Hunger: standardFoodHunger},
	FoodSandwich:   {Name: "Sandwich", Hunger: 50},
	FoodCake:       {Name: "Cake", Hunger: 10, Happiness: 15},
	FoodCoffee:     {Name: "Coffee", Verb: "Drink", Energy: 35},
	FoodEnergyPill: {Name: "Energy Pill", Verb: "Use", Energy: 100, WakesUp: true},
}

var (
	ErrPetAway = errors.New("pet has run away")
	ErrAsleep  = errors.New("pet is asleep")
	ErrNoFood  = errors.New("no food of that kind in inventory")
)

// AddFood puts n items of a kind into the inventory (e.g. mini-game rewards).
func (p *Pet) AddFood(kind FoodKind, n int) {
	if p.Inventory == nil {
		p.Inventory = map[FoodKind]int{}
	}
	p.Inventory[kind] += n
}

// FoodCount returns how many of a kind are held.
func (p *Pet) FoodCount(kind FoodKind) int { return p.Inventory[kind] }

// Feed consumes one item and applies its effect. Feeding is inventory-driven —
// it is not a direct hunger fill. A sleeping pet refuses everything except
// WakesUp items (the energy pill), which end the nap on the spot.
func (p *Pet) Feed(kind FoodKind) error {
	if p.Away {
		return ErrPetAway
	}
	def := Foods[kind]
	if p.Asleep && !def.WakesUp {
		return ErrAsleep
	}
	if p.Inventory[kind] <= 0 {
		return ErrNoFood
	}
	p.Inventory[kind]--
	p.Stats.Hunger = clamp(p.Stats.Hunger + def.Hunger)
	p.Stats.Happiness = clamp(p.Stats.Happiness + def.Happiness)
	p.Stats.Energy = clamp(p.Stats.Energy + def.Energy)
	if p.Asleep && def.WakesUp {
		p.Asleep = false
	}
	return nil
}
