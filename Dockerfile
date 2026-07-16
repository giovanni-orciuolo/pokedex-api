FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/pokedex ./cmd/server

FROM gcr.io/distroless/static-debian12
COPY --from=build /bin/pokedex /pokedex
EXPOSE 5000
ENTRYPOINT ["/pokedex"]
