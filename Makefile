# Makefile - SublimeGo
# Multi-plateforme: Windows, Linux, macOS
# Sans dependance Node.js ou npm

PROJECT_NAME := SublimeGo

.PHONY: all build run test clean lint css css-watch tailwind-download generate dev

all: build

# Installation des outils Go
install-tools:
	go install github.com/air-verse/air@latest
	go install github.com/a-h/templ/cmd/templ@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Telecharger Tailwind CSS Standalone CLI (une seule fois)
tailwind-download:
	go run ./cmd/tools/tailwind download

# Compilation CSS avec Tailwind (production - minifie)
css:
	go run ./cmd/tools/tailwind build

# Compilation CSS en mode watch (developpement)
css-watch:
	go run ./cmd/tools/tailwind watch

# Generation du code (Templ)
generate:
	templ generate

# Lancer en mode developpement (Hot Reload)
dev: generate
	air

# Construire le binaire final (Production)
build: generate css
	go build -o bin/$(PROJECT_NAME) ./cmd/sublimego

# Verifier la qualite du code
lint:
	golangci-lint run

# Nettoyer
clean:
	rm -rf bin
	rm -rf tmp
	rm -rf tools