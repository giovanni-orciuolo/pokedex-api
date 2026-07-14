package pokemon

import "errors"

var ErrNotFound = errors.New("pokemon not found")

type Pokemon struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Habitat     string `json:"habitat"`
	IsLegendary bool   `json:"isLegendary"`
}
