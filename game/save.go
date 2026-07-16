package game

import (
	"encoding/json"

	"sapootchi/simulation"
)

// Settings are player preferences, persisted alongside the pet.
type Settings struct {
	// RealSpriteInGames renders the actual blob sprite as the player character
	// inside mini-games instead of the simple shape stand-in.
	RealSpriteInGames bool
}

// saveFile is the on-disk format: pet + settings in one JSON document.
type saveFile struct {
	Pet      *simulation.Pet
	Settings Settings
}

func encodeSave(pet *simulation.Pet, settings Settings) ([]byte, error) {
	return json.Marshal(saveFile{Pet: pet, Settings: settings})
}

// decodeSave reads the current format, falling back to the legacy format that
// was a bare simulation.Pet document.
func decodeSave(data []byte) (*simulation.Pet, Settings, error) {
	var sf saveFile
	if err := json.Unmarshal(data, &sf); err == nil && sf.Pet != nil {
		return sf.Pet, sf.Settings, nil
	}
	pet, err := simulation.Load(data) // legacy: bare pet
	if err != nil {
		return nil, Settings{}, err
	}
	return pet, Settings{}, nil
}
