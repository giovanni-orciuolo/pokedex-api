# Pokédex API

A small REST API that serves Pokémon information, with an optional "fun" translation of the
Pokémon's description (Yoda for legendary or cave-dwelling Pokémon, Shakespeare for everyone
else).

Built in Go with **zero external dependencies**, only the standard library.

## Endpoints

| Method | Path                          | Description                                    |
|--------|-------------------------------|------------------------------------------------|
| GET    | `/pokemon/{name}`             | Basic Pokémon information                      |
| GET    | `/pokemon/translated/{name}`  | Same, with a fun translation of the description |

Example:

```
$ curl http://localhost:5000/pokemon/mewtwo
{"name":"mewtwo","description":"It was created by a scientist after years of horrific gene splicing and DNA engineering experiments.","habitat":"rare","isLegendary":true}

$ curl http://localhost:5000/pokemon/translated/mewtwo
{"name":"mewtwo","description":"Created by a scientist after years of horrific gene splicing and dna engineering experiments, it was.","habitat":"rare","isLegendary":true}
```

Error responses are JSON as well: `404` when the Pokémon does not exist, `502` when PokeAPI is
unreachable or misbehaving.

## How to run

### With Docker (nothing else required)

Install Docker: https://docs.docker.com/get-docker/

```
docker build -t pokedex .
docker run --rm -p 5000:5000 pokedex
```

### From source

Install Go 1.22 or newer: https://go.dev/doc/install

```
go run ./cmd/server
```

The server listens on `:5000` by default.

### Run the tests

```
go test ./...
```

### Configuration

Everything is configurable via environment variables (useful for tests, local mirrors, or
pointing at a paid FunTranslations plan):

| Variable              | Default                              |
|-----------------------|--------------------------------------|
| `ADDR`                | `:5000`                              |
| `POKEAPI_URL`         | `https://pokeapi.co`                 |
| `FUNTRANSLATIONS_URL` | `https://api.funtranslations.mercxry.me/v1` |

## Project layout

```
cmd/server/          entrypoint: config, wiring, graceful shutdown
internal/pokemon/    domain type shared by all layers
internal/pokeapi/    HTTP client for PokeAPI (pokemon-species)
internal/api/        HTTP handlers, routing, translation-selection rule
internal/translator/ HTTP client for FunTranslations (yoda/shakespeare)
```

The handlers depend on small interfaces (`SpeciesFetcher`, `Translator`) rather than on the
concrete clients, so the HTTP layer is unit-tested with in-memory stubs and each client is
tested against a fake upstream (`httptest.Server`).

## Design decisions

- **No external dependencies.** Go 1.22's `net/http.ServeMux` supports method + path-parameter
  routing, which is all this service needs. Fewer dependencies means a smaller supply-chain
  surface and nothing to keep patched.
- **Translation failures degrade gracefully.** The FunTranslations mirror is rate-limited
  (5 requests/minute, communicated via a 429 with `retry_after` in the body); per the
  requirements, any translation failure falls back to the standard description and the
  endpoint still returns `200`. The failure is logged.
- **The mirror's contract, not the official docs.** The challenge URL serves the mirror's
  Swagger page; the actual API (per its OpenAPI spec) lives at
  `api.funtranslations.mercxry.me/v1`, is POST-only, and uses `/translate/{style}` without
  the `.json` suffix used by the official funtranslations.com API.
- **Habitat can be missing.** Newer-generation Pokemon have `habitat: null` in PokeAPI; the API
  returns `"unknown"` rather than an empty string so the field is always meaningful.
- **Flavor text is normalized.** PokeAPI descriptions contain raw `\n`/`\f` control characters
  from the original game data; these are collapsed into single spaces.
- **First English flavor text is used**, as permitted by the task ("you can use any of the
  English descriptions").

## What I would do differently for production

- **Caching.** Pokemon data is effectively immutable, so responses from PokeAPI (and
  translations) should be cached aggressively: an in-memory LRU with TTL for a single
  instance, or Redis for a fleet. This is the single highest-impact change given the
  FunTranslations rate limit.
- **Resilience.** Retries with backoff and jitter for transient upstream failures, a circuit
  breaker per upstream, and tighter per-request timeouts/budgets.
- **Observability.** Structured logging (`log/slog`), request metrics (latency, status codes,
  upstream error rates), tracing across the two upstream calls, and a `/health` endpoint for
  orchestrator probes.
- **API hygiene.** Version the API (`/v1/pokemon/...`), publish an OpenAPI spec, and add
  request-level rate limiting.
- **Fallback transparency.** When translation fails we silently return the standard
  description; in production I would surface this (e.g. a `translationApplied` field or a
  response header) so clients can distinguish the two cases. Kept the response minimal here to
  match the specified contract.
- **Config.** Environment variables are fine at this size; with more knobs I would introduce a
  typed config struct validated at startup.
- **CI.** Lint (`golangci-lint`), vet, tests, and image build on every push.
