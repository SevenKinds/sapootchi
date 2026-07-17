package game

import (
	"encoding/json"

	"sapootchi/simulation"
)

// Settings are player preferences, persisted alongside the pets.
type Settings struct {
	// RealSpriteInGames renders the actual blob sprite as the player character
	// inside mini-games instead of the simple shape stand-in.
	RealSpriteInGames bool
	// Skin is LEGACY: skins are per-pet now (Pet.Skin). Kept for decoding old
	// saves; migrated and cleared on load.
	Skin string `json:",omitempty"`
	// OwnedSkins are the looks unlocked in the shop (account-wide: buy once,
	// any pet can wear it). Classic is always owned.
	OwnedSkins []string
	// Theme is the app palette name ("" = the first/default theme).
	Theme string
}

// saveFile is the on-disk format. Multi-pet: Pets + Active. Older formats are
// migrated on load (single Pet field, or a bare simulation.Pet document).
type saveFile struct {
	Pet      *simulation.Pet `json:",omitempty"` // legacy single-pet field
	Pets     []*simulation.Pet
	Active   int
	Settings Settings
}

func encodeSave(pets []*simulation.Pet, active int, settings Settings) ([]byte, error) {
	return json.Marshal(saveFile{Pets: pets, Active: active, Settings: settings})
}

// decodeSave reads any save format generation and returns the pet roster.
func decodeSave(data []byte) (pets []*simulation.Pet, active int, s Settings, err error) {
	var sf saveFile
	if err := json.Unmarshal(data, &sf); err == nil {
		if len(sf.Pets) > 0 {
			a := sf.Active
			if a < 0 || a >= len(sf.Pets) {
				a = 0
			}
			return migrateSkin(sf.Pets, a, sf.Settings)
		}
		if sf.Pet != nil { // single-pet format
			return migrateSkin([]*simulation.Pet{sf.Pet}, 0, sf.Settings)
		}
	}
	pet, err := simulation.Load(data) // oldest: bare pet
	if err != nil {
		return nil, 0, Settings{}, err
	}
	return []*simulation.Pet{pet}, 0, Settings{}, nil
}

// migrateSkin moves the legacy global Settings.Skin onto the pets (they all
// shared one look before skins became per-pet), and grandfathers any skin a
// pet already wears into the owned set (they predate the shop unlock system).
func migrateSkin(pets []*simulation.Pet, active int, s Settings) ([]*simulation.Pet, int, Settings, error) {
	if s.Skin != "" {
		for _, p := range pets {
			if p.Skin == "" {
				p.Skin = s.Skin
			}
		}
		s.Skin = ""
	}
	owned := map[string]bool{}
	for _, n := range s.OwnedSkins {
		owned[n] = true
	}
	for _, p := range pets {
		if p.Skin != "" && !owned[p.Skin] {
			owned[p.Skin] = true
			s.OwnedSkins = append(s.OwnedSkins, p.Skin)
		}
	}
	return pets, active, s, nil
}
