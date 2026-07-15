package translator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type Style string

const (
	Yoda        Style = "yoda"
	Shakespeare Style = "shakespeare"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string, httpClient *http.Client) *Client {
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), httpClient: httpClient}
}

type translationResponse struct {
	Contents struct {
		Translated string `json:"translated"`
	} `json:"contents"`
}

func (c *Client) Translate(ctx context.Context, text string, style Style) (string, error) {
	endpoint := fmt.Sprintf("%s/translate/%s.json", c.baseURL, style)
	form := url.Values{"text": {text}}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("calling funtranslations: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("funtranslations returned status %d", resp.StatusCode)
	}

	var translation translationResponse
	if err := json.NewDecoder(resp.Body).Decode(&translation); err != nil {
		return "", fmt.Errorf("decoding funtranslations response: %w", err)
	}
	if translation.Contents.Translated == "" {
		return "", errors.New("funtranslations returned an empty translation")
	}

	return translation.Contents.Translated, nil
}
