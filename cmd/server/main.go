package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pokedex/internal/api"
	"pokedex/internal/pokeapi"
	"pokedex/internal/translator"
)

func main() {
	httpClient := &http.Client{Timeout: 10 * time.Second}

	handler := api.NewHandler(
		pokeapi.NewClient(envOr("POKEAPI_URL", "https://pokeapi.co"), httpClient),
		translator.NewClient(envOr("FUNTRANSLATIONS_URL", "https://api.funtranslations.mercxry.me/v1"), httpClient),
	)

	server := &http.Server{
		Addr:              envOr("ADDR", ":5000"),
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("pokedex listening on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
