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
)

// FoodDef is the effect of eating one item.
type FoodDef struct {
	Name      string
	Hunger    float64
	Happiness float64
}

// Foods is the item catalog. Apple is the standard (+25%).
var Foods = map[FoodKind]FoodDef{
	FoodApple:    {Name: "Apple", Hunger: standardFoodHunger, Happiness: 0},
	FoodSandwich: {Name: "Sandwich", Hunger: 50, Happiness: 0},
	FoodCake:     {Name: "Cake", Hunger: 10, Happiness: 15},
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
// it is not a direct hunger fill.
func (p *Pet) Feed(kind FoodKind) error {
	if p.Away {
		return ErrPetAway
	}
	if p.Asleep {
		return ErrAsleep
	}
	if p.Inventory[kind] <= 0 {
		return ErrNoFood
	}
	p.Inventory[kind]--
	def := Foods[kind]
	p.Stats.Hunger = clamp(p.Stats.Hunger + def.Hunger)
	p.Stats.Happiness = clamp(p.Stats.Happiness + def.Happiness)
	return nil
}
