package game

import (
	"encoding/json"
	"testing"

	"sapootchi/simulation"
)

// Coins moved from per-pet to one shared wallet; legacy saves must fold each
// pet's balance into Settings.Coins and zero the pets, idempotently.
func TestCoinsMigrateToSharedWallet(t *testing.T) {
	data, err := json.Marshal(saveFile{
		Pets:   []*simulation.Pet{{Coins: 50}, {Coins: 30}},
		Active: 0,
	})
	if err != nil {
		t.Fatal(err)
	}

	pets, _, s, err := decodeSave(data)
	if err != nil {
		t.Fatal(err)
	}
	if s.Coins != 80 {
		t.Fatalf("shared wallet = %d, want 80", s.Coins)
	}
	for i, p := range pets {
		if p.Coins != 0 {
			t.Fatalf("pet %d still holds %d coins, want 0", i, p.Coins)
		}
	}

	// Re-saving the migrated state and loading again must not change the total.
	data2, err := encodeSave(pets, 0, s)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, s2, _ := decodeSave(data2); s2.Coins != 80 {
		t.Fatalf("after round-trip shared wallet = %d, want 80", s2.Coins)
	}
}
