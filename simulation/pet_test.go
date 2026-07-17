package simulation

import (
	"math/rand"
	"testing"
	"time"
)

func newTestPet() (*Pet, time.Time) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	return NewPet("Test", PersonalityLazy, now), now
}

func TestHungerDecays25PercentPerDay(t *testing.T) {
	p, now := newTestPet()
	p.Stats.Hunger = 100
	p.Update(now.Add(24*time.Hour), rand.New(rand.NewSource(1)))
	if got := p.Stats.Hunger; got < 74.9 || got > 75.1 {
		t.Fatalf("hunger after 1 day = %.2f, want ~75", got)
	}
}

func TestFullToEmptyInFourDays(t *testing.T) {
	p, now := newTestPet()
	p.Stats.Hunger = 100
	// Clamp caps a single absence at 48h, so step day-by-day.
	for i := 1; i <= 4; i++ {
		p.Update(now.Add(time.Duration(i)*24*time.Hour), rand.New(rand.NewSource(1)))
	}
	if p.Stats.Hunger > 0 {
		t.Fatalf("hunger after 4 days = %.2f, want 0", p.Stats.Hunger)
	}
}

func TestStandardFoodRestores25(t *testing.T) {
	p, _ := newTestPet()
	p.Stats.Hunger = 0
	p.Inventory = map[FoodKind]int{FoodApple: 1}
	if err := p.Feed(FoodApple); err != nil {
		t.Fatal(err)
	}
	if got := p.Stats.Hunger; got < 24.9 || got > 25.1 {
		t.Fatalf("hunger after 1 apple = %.2f, want ~25", got)
	}
	if p.FoodCount(FoodApple) != 0 {
		t.Fatalf("apple not consumed")
	}
}

func TestFeedWithoutFoodErrors(t *testing.T) {
	p, _ := newTestPet()
	p.Inventory = map[FoodKind]int{}
	if err := p.Feed(FoodApple); err != ErrNoFood {
		t.Fatalf("want ErrNoFood, got %v", err)
	}
}

func TestRunsAwayWhenHungerHitsZero(t *testing.T) {
	p, now := newTestPet()
	p.Stats.Hunger = 10
	// 2 days at 25%/day would go negative -> clamps to 0 -> runs away.
	p.Update(now.Add(48*time.Hour), rand.New(rand.NewSource(1)))
	if !p.Away {
		t.Fatalf("pet should have run away when hunger hit 0")
	}
}

func TestNoReturnRollsUntilFoodLeftOut(t *testing.T) {
	p, now := newTestPet()
	p.Away = true
	p.LastSeen = now
	// Many days pass but no food out -> never returns.
	p.Update(now.Add(100*24*time.Hour), rand.New(rand.NewSource(1)))
	if !p.Away {
		t.Fatalf("pet must not return before food is left out")
	}
}

func TestReturnsEventuallyWithFoodOut(t *testing.T) {
	p, now := newTestPet()
	p.Away = true
	p.LastSeen = now
	p.LeaveFoodOut()
	// With 28%/day, 100 days is overwhelmingly likely to bring him home.
	p.Update(now.Add(100*24*time.Hour), rand.New(rand.NewSource(1)))
	if p.Away {
		t.Fatalf("pet should have returned within 100 days of food being out")
	}
	if p.Stats.Hunger <= 0 {
		t.Fatalf("returned pet should have eaten the food left out")
	}
}

func TestReturnRateIsRoughly28Percent(t *testing.T) {
	const trials = 20000
	rng := rand.New(rand.NewSource(42))
	returned := 0
	for i := 0; i < trials; i++ {
		p, now := newTestPet()
		p.Away = true
		p.LastSeen = now
		p.LeaveFoodOut()
		// Exactly one day elapses -> exactly one roll.
		p.Update(now.Add(24*time.Hour), rng)
		if !p.Away {
			returned++
		}
	}
	rate := float64(returned) / float64(trials)
	if rate < 0.26 || rate > 0.30 {
		t.Fatalf("single-day return rate = %.3f, want ~0.28", rate)
	}
}

func TestCoffeeRestoresEnergy(t *testing.T) {
	p, _ := newTestPet()
	p.Stats.Energy = 40
	hungerBefore := p.Stats.Hunger
	p.Inventory = map[FoodKind]int{FoodCoffee: 1}
	if err := p.Feed(FoodCoffee); err != nil {
		t.Fatal(err)
	}
	if got := p.Stats.Energy; got != 75 {
		t.Fatalf("energy after coffee = %.2f, want 75", got)
	}
	if p.Stats.Hunger != hungerBefore {
		t.Fatalf("coffee should not affect hunger, got %.2f", p.Stats.Hunger)
	}
}

func TestRescueStart(t *testing.T) {
	p, _ := newTestPet()
	// Found in the nick of time: hungry and grubby but awake and recoverable.
	if p.Stats.Hunger > 25 || p.Stats.Hygiene > 30 || p.Stats.Happiness > 40 {
		t.Fatalf("new pet should start in rescue shape, got %+v", p.Stats)
	}
	if p.Asleep || p.Away {
		t.Fatalf("new pet must be awake and present")
	}
	// The starter apples must be enough to pull hunger out of danger.
	p.Feed(FoodApple)
	p.Feed(FoodApple)
	if p.Stats.Hunger < 50 {
		t.Fatalf("starter apples should rescue hunger to ~62, got %.2f", p.Stats.Hunger)
	}
}

func TestMoodPrecedence(t *testing.T) {
	p, _ := newTestPet()
	p.Stats = Stats{Happiness: 10, Hunger: 10, Hygiene: 100, Energy: 10}
	if p.Mood() != MoodHungry {
		t.Fatalf("hunger should win precedence, got %v", p.Mood())
	}
	// Full energy = energized = Excited (energy is inverted).
	p.Stats = Stats{Happiness: 100, Hunger: 100, Hygiene: 100, Energy: 100}
	if p.Mood() != MoodExcited {
		t.Fatalf("full energy should be Excited, got %v", p.Mood())
	}
	// High happiness but mid energy = Happy.
	p.Stats = Stats{Happiness: 100, Hunger: 100, Hygiene: 100, Energy: 70}
	if p.Mood() != MoodHappy {
		t.Fatalf("high happiness, mid energy should be Happy, got %v", p.Mood())
	}
}

func TestEnergyRegensWhileIdle(t *testing.T) {
	p, now := newTestPet()
	p.Stats.Energy = 30
	// Half a day idle: +regen/2, still under the 100 clamp.
	p.Update(now.Add(12*time.Hour), rand.New(rand.NewSource(1)))
	want := 30 + energyIdleRegenPerDay/2
	if got := p.Stats.Energy; got < want-0.1 || got > want+0.1 {
		t.Fatalf("energy after half an idle day = %.2f, want ~%.0f", got, want)
	}
}

func TestEnergizedWhenFull(t *testing.T) {
	p, now := newTestPet()
	p.Stats.Energy = 100
	p.Update(now.Add(24*time.Hour), rand.New(rand.NewSource(1)))
	if !p.Energized() {
		t.Fatalf("full-energy pet should be energized")
	}
	if p.Stats.Energy != 100 {
		t.Fatalf("energy should clamp at 100, got %.2f", p.Stats.Energy)
	}
}

func TestFallsAsleepAtZeroAndWakesAt50(t *testing.T) {
	p, now := newTestPet()
	// Simulate a mini-game draining energy to 0.
	p.Stats.Energy = 0
	// One tick trips the sleep trigger.
	now = now.Add(time.Second)
	p.Update(now, rand.New(rand.NewSource(1)))
	if !p.Asleep {
		t.Fatalf("pet should be asleep at energy 0")
	}
	// It should stay asleep before reaching the wake threshold...
	p.Stats.Energy = 40
	now = now.Add(time.Second)
	p.Update(now, rand.New(rand.NewSource(1)))
	if !p.Asleep {
		t.Fatalf("pet should stay asleep below wake threshold")
	}
	// ...and wake once regenerated past it.
	p.Stats.Energy = 50
	now = now.Add(time.Second)
	p.Update(now, rand.New(rand.NewSource(1)))
	if p.Asleep {
		t.Fatalf("pet should wake at/after energy %.0f", energyWakeThreshold)
	}
}

func TestEnergyPillWakesSleepingPet(t *testing.T) {
	p, _ := newTestPet()
	p.Asleep = true
	p.Stats.Energy = 0
	p.Inventory = map[FoodKind]int{FoodEnergyPill: 1, FoodCoffee: 1}
	// Coffee must still be refused — the nap is sacred to everything else.
	if err := p.Feed(FoodCoffee); err != ErrAsleep {
		t.Fatalf("coffee while asleep should ErrAsleep, got %v", err)
	}
	// The pill overrides sleep: that is its whole point.
	if err := p.Feed(FoodEnergyPill); err != nil {
		t.Fatal(err)
	}
	if p.Asleep {
		t.Fatalf("energy pill should wake the pet")
	}
	if p.Stats.Energy != 100 {
		t.Fatalf("energy after pill = %.2f, want 100", p.Stats.Energy)
	}
}

func TestRestIsVoluntaryNapUntilFull(t *testing.T) {
	p, now := newTestPet()
	p.Stats.Energy = 60
	p.Rest()
	if !p.Asleep || !p.NapVoluntary {
		t.Fatalf("rest should start a voluntary nap")
	}
	// Past the forced-nap threshold (50) but not full -> STAYS asleep.
	p.Stats.Energy = 80
	now = now.Add(time.Second)
	p.Update(now, rand.New(rand.NewSource(1)))
	if !p.Asleep {
		t.Fatalf("voluntary nap must not wake at the forced threshold")
	}
	// Sleep regen carries it to 100 -> wakes, flag cleared.
	now = now.Add(2 * time.Hour)
	p.Update(now, rand.New(rand.NewSource(1)))
	if p.Asleep || p.NapVoluntary {
		t.Fatalf("voluntary nap should end at full energy, got %.1f asleep=%v", p.Stats.Energy, p.Asleep)
	}
	if p.Stats.Energy < 99.9 {
		t.Fatalf("should wake full, got %.1f", p.Stats.Energy)
	}
}

func TestPettingCooldown(t *testing.T) {
	p, now := newTestPet()
	before := p.Stats.Happiness
	if !p.Pet(now) {
		t.Fatalf("first pet should grant happiness")
	}
	if p.Stats.Happiness <= before {
		t.Fatalf("happiness should rise on first pet")
	}
	mid := p.Stats.Happiness
	if p.Pet(now.Add(30 * time.Second)) {
		t.Fatalf("petting inside the cooldown must not grant again")
	}
	if p.Stats.Happiness != mid {
		t.Fatalf("spam petting changed happiness")
	}
	if !p.Pet(now.Add(4 * time.Minute)) {
		t.Fatalf("petting after the cooldown should grant again")
	}
}

func TestCannotActWhileAsleep(t *testing.T) {
	p, _ := newTestPet()
	p.Asleep = true
	p.Inventory = map[FoodKind]int{FoodApple: 1}
	if err := p.Feed(FoodApple); err != ErrAsleep {
		t.Fatalf("feeding asleep pet should error, got %v", err)
	}
	before := p.Stats.Hygiene
	p.Bathe()
	if p.Stats.Hygiene != before {
		t.Fatalf("bathing an asleep pet should be a no-op")
	}
}

func TestOfflineDecayClamped(t *testing.T) {
	p, now := newTestPet()
	p.Stats.Happiness = 100
	// 10 days away but clamp caps at 48h => at most 2 days of happiness decay.
	p.Update(now.Add(10*24*time.Hour), rand.New(rand.NewSource(1)))
	if p.Stats.Happiness < 100-2*happinessDecayPerDay-0.1 {
		t.Fatalf("offline decay not clamped: happiness=%.2f", p.Stats.Happiness)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	p, _ := newTestPet()
	p.Coins = 123
	p.Skin = "halloween"
	p.AddFood(FoodCake, 3)
	data, err := Save(p)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Load(data)
	if err != nil {
		t.Fatal(err)
	}
	if got.Coins != 123 || got.FoodCount(FoodCake) != 3 || got.Skin != "halloween" {
		t.Fatalf("round trip mismatch: coins=%d cake=%d skin=%q",
			got.Coins, got.FoodCount(FoodCake), got.Skin)
	}
}

func TestEvolveStubAfterAge(t *testing.T) {
	p, now := newTestPet()
	if p.CanEvolve(now) {
		t.Fatalf("baby should not evolve immediately")
	}
	later := now.Add(evolveAge + time.Hour)
	if !p.CanEvolve(later) {
		t.Fatalf("baby should be able to evolve after evolveAge")
	}
	p.Hidden.Intelligence = 70
	p.Evolve(later)
	if p.Phase != PhaseSmart {
		t.Fatalf("high intelligence should branch to Smart, got %v", p.Phase)
	}
}
