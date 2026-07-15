package api

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"pokedex/internal/pokemon"
	"pokedex/internal/translator"
)

type SpeciesFetcher interface {
	Species(ctx context.Context, name string) (pokemon.Pokemon, error)
}

type Translator interface {
	Translate(ctx context.Context, text string, style translator.Style) (string, error)
}

type handler struct {
	species    SpeciesFetcher
	translator Translator
}

func NewHandler(species SpeciesFetcher, translator Translator) http.Handler {
	h := &handler{species: species, translator: translator}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /pokemon/{name}", h.basic)
	mux.HandleFunc("GET /pokemon/translated/{name}", h.translated)
	return mux
}

func (h *handler) basic(w http.ResponseWriter, r *http.Request) {
	p, ok := h.fetch(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (h *handler) translated(w http.ResponseWriter, r *http.Request) {
	p, ok := h.fetch(w, r)
	if !ok {
		return
	}

	translated, err := h.translator.Translate(r.Context(), p.Description, styleFor(p))
	if err != nil {
		log.Printf("translating %q failed, using standard description: %v", p.Name, err)
	} else {
		p.Description = translated
	}
	writeJSON(w, http.StatusOK, p)
}

func (h *handler) fetch(w http.ResponseWriter, r *http.Request) (pokemon.Pokemon, bool) {
	name := r.PathValue("name")

	p, err := h.species.Species(r.Context(), name)
	if errors.Is(err, pokemon.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "pokemon not found"})
		return pokemon.Pokemon{}, false
	}
	if err != nil {
		log.Printf("fetching species %q: %v", name, err)
		writeJSON(w, http.StatusBadGateway, errorResponse{Error: "upstream service unavailable"})
		return pokemon.Pokemon{}, false
	}
	return p, true
}

func styleFor(p pokemon.Pokemon) translator.Style {
	if p.Habitat == "cave" || p.IsLegendary {
		return translator.Yoda
	}
	return translator.Shakespeare
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("encoding response: %v", err)
	}
}
