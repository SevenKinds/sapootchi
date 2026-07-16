package simulation

import "encoding/json"

// Save serializes the pet to JSON. Persistence I/O (file, localStorage) is the
// caller's job — this package stays free of platform concerns.
func Save(p *Pet) ([]byte, error) {
	return json.Marshal(p)
}

// Load deserializes a pet from JSON produced by Save.
func Load(data []byte) (*Pet, error) {
	var p Pet
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	if p.Inventory == nil {
		p.Inventory = map[FoodKind]int{}
	}
	return &p, nil
}
