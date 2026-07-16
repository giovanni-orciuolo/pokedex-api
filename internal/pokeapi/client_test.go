package pokeapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"pokedex/internal/pokemon"
)

const mewtwoJSON = `{
	"name": "mewtwo",
	"is_legendary": true,
	"habitat": {"name": "rare"},
	"flavor_text_entries": [
		{"flavor_text": "Es wurde von einem Forscher erzeugt.", "language": {"name": "de"}},
		{"flavor_text": "It was created by\na scientist after\fyears of experiments.", "language": {"name": "en"}}
	]
}`

func TestSpeciesMapsResponseToDomainType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/pokemon-species/mewtwo" {
			t.Errorf("path = %q, want /api/v2/pokemon-species/mewtwo", r.URL.Path)
		}
		w.Write([]byte(mewtwoJSON))
	}))
	defer server.Close()

	got, err := NewClient(server.URL, server.Client()).Species(context.Background(), "MEWTWO")
	if err != nil {
		t.Fatal(err)
	}

	want := pokemon.Pokemon{
		Name:        "mewtwo",
		Description: "It was created by a scientist after years of experiments.",
		Habitat:     "rare",
		IsLegendary: true,
	}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestSpeciesDefaultsHabitatToUnknownWhenMissing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"name": "arceus", "is_legendary": false, "habitat": null, "flavor_text_entries": []}`))
	}))
	defer server.Close()

	got, err := NewClient(server.URL, server.Client()).Species(context.Background(), "arceus")
	if err != nil {
		t.Fatal(err)
	}
	if got.Habitat != "unknown" {
		t.Errorf("habitat = %q, want %q", got.Habitat, "unknown")
	}
}

func TestSpeciesReturnsErrNotFoundOn404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	_, err := NewClient(server.URL, server.Client()).Species(context.Background(), "agumon")
	if !errors.Is(err, pokemon.ErrNotFound) {
		t.Errorf("err = %v, want pokemon.ErrNotFound", err)
	}
}

func TestSpeciesReturnsErrorOnServerFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	_, err := NewClient(server.URL, server.Client()).Species(context.Background(), "mewtwo")
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
}
