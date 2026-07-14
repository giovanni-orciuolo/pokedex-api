package pokeapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"pokedex/internal/pokemon"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string, httpClient *http.Client) *Client {
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), httpClient: httpClient}
}

type speciesResponse struct {
	Name        string `json:"name"`
	IsLegendary bool   `json:"is_legendary"`
	Habitat     *struct {
		Name string `json:"name"`
	} `json:"habitat"`
	FlavorTextEntries []struct {
		FlavorText string `json:"flavor_text"`
		Language   struct {
			Name string `json:"name"`
		} `json:"language"`
	} `json:"flavor_text_entries"`
}

func (c *Client) Species(ctx context.Context, name string) (pokemon.Pokemon, error) {
	endpoint := fmt.Sprintf("%s/api/v2/pokemon-species/%s",
		c.baseURL, url.PathEscape(strings.ToLower(name)))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return pokemon.Pokemon{}, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return pokemon.Pokemon{}, fmt.Errorf("calling pokeapi: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return pokemon.Pokemon{}, pokemon.ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return pokemon.Pokemon{}, fmt.Errorf("pokeapi returned status %d", resp.StatusCode)
	}

	var species speciesResponse
	if err := json.NewDecoder(resp.Body).Decode(&species); err != nil {
		return pokemon.Pokemon{}, fmt.Errorf("decoding pokeapi response: %w", err)
	}

	return pokemon.Pokemon{
		Name:        species.Name,
		Description: englishDescription(species),
		Habitat:     habitatName(species),
		IsLegendary: species.IsLegendary,
	}, nil
}

func englishDescription(s speciesResponse) string {
	for _, entry := range s.FlavorTextEntries {
		if entry.Language.Name == "en" {
			return entry.FlavorText
		}
	}
	return ""
}

func habitatName(s speciesResponse) string {
	if s.Habitat == nil {
		return "unknown"
	}
	return s.Habitat.Name
}
