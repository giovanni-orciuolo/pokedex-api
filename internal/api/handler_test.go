package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"pokedex/internal/pokemon"
	"pokedex/internal/translator"
)

type stubSpecies struct {
	result pokemon.Pokemon
	err    error
}

func (s stubSpecies) Species(_ context.Context, _ string) (pokemon.Pokemon, error) {
	return s.result, s.err
}

type spyTranslator struct {
	result    string
	err       error
	gotText   string
	gotStyle  translator.Style
	callCount int
}

func (s *spyTranslator) Translate(_ context.Context, text string, style translator.Style) (string, error) {
	s.callCount++
	s.gotText = text
	s.gotStyle = style
	return s.result, s.err
}

func TestBasicEndpointReturnsPokemon(t *testing.T) {
	mewtwo := pokemon.Pokemon{
		Name:        "mewtwo",
		Description: "It was created by a scientist.",
		Habitat:     "rare",
		IsLegendary: true,
	}
	h := NewHandler(stubSpecies{result: mewtwo}, &spyTranslator{})

	resp := serve(t, h, "/pokemon/mewtwo")

	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.Code, http.StatusOK)
	}
	if got := decodePokemon(t, resp); got != mewtwo {
		t.Errorf("body = %+v, want %+v", got, mewtwo)
	}
}

func TestBasicEndpointReturns404WhenPokemonDoesNotExist(t *testing.T) {
	h := NewHandler(stubSpecies{err: pokemon.ErrNotFound}, &spyTranslator{})

	resp := serve(t, h, "/pokemon/agumon")

	if resp.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", resp.Code, http.StatusNotFound)
	}
}

func TestBasicEndpointReturns502WhenPokeAPIFails(t *testing.T) {
	h := NewHandler(stubSpecies{err: errors.New("boom")}, &spyTranslator{})

	resp := serve(t, h, "/pokemon/mewtwo")

	if resp.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d", resp.Code, http.StatusBadGateway)
	}
}

func TestTranslatedEndpointUsesYodaForLegendaryPokemon(t *testing.T) {
	legendary := pokemon.Pokemon{Name: "mewtwo", Description: "original", Habitat: "rare", IsLegendary: true}
	spy := &spyTranslator{result: "translated, it was"}
	h := NewHandler(stubSpecies{result: legendary}, spy)

	resp := serve(t, h, "/pokemon/translated/mewtwo")

	if spy.gotStyle != translator.Yoda {
		t.Errorf("style = %q, want %q", spy.gotStyle, translator.Yoda)
	}
	if got := decodePokemon(t, resp); got.Description != "translated, it was" {
		t.Errorf("description = %q, want the translated one", got.Description)
	}
}

func TestTranslatedEndpointUsesYodaForCavePokemon(t *testing.T) {
	cave := pokemon.Pokemon{Name: "zubat", Description: "original", Habitat: "cave", IsLegendary: false}
	spy := &spyTranslator{result: "translated"}
	h := NewHandler(stubSpecies{result: cave}, spy)

	serve(t, h, "/pokemon/translated/zubat")

	if spy.gotStyle != translator.Yoda {
		t.Errorf("style = %q, want %q", spy.gotStyle, translator.Yoda)
	}
}

func TestTranslatedEndpointUsesShakespeareForEveryoneElse(t *testing.T) {
	regular := pokemon.Pokemon{Name: "pikachu", Description: "original", Habitat: "forest", IsLegendary: false}
	spy := &spyTranslator{result: "translated"}
	h := NewHandler(stubSpecies{result: regular}, spy)

	serve(t, h, "/pokemon/translated/pikachu")

	if spy.gotStyle != translator.Shakespeare {
		t.Errorf("style = %q, want %q", spy.gotStyle, translator.Shakespeare)
	}
	if spy.gotText != "original" {
		t.Errorf("translated text = %q, want the standard description", spy.gotText)
	}
}

func TestTranslatedEndpointFallsBackToStandardDescriptionOnTranslationFailure(t *testing.T) {
	regular := pokemon.Pokemon{Name: "pikachu", Description: "original", Habitat: "forest", IsLegendary: false}
	spy := &spyTranslator{err: errors.New("rate limited")}
	h := NewHandler(stubSpecies{result: regular}, spy)

	resp := serve(t, h, "/pokemon/translated/pikachu")

	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.Code, http.StatusOK)
	}
	if got := decodePokemon(t, resp); got.Description != "original" {
		t.Errorf("description = %q, want fallback to %q", got.Description, "original")
	}
}

func TestBasicEndpointNeverCallsTheTranslator(t *testing.T) {
	spy := &spyTranslator{result: "translated"}
	h := NewHandler(stubSpecies{result: pokemon.Pokemon{Name: "pikachu"}}, spy)

	serve(t, h, "/pokemon/pikachu")

	if spy.callCount != 0 {
		t.Errorf("translator called %d times, want 0", spy.callCount)
	}
}

func serve(t *testing.T, h http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	resp := httptest.NewRecorder()
	h.ServeHTTP(resp, req)
	return resp
}

func decodePokemon(t *testing.T, resp *httptest.ResponseRecorder) pokemon.Pokemon {
	t.Helper()
	var p pokemon.Pokemon
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		t.Fatalf("decoding response body: %v", err)
	}
	return p
}
