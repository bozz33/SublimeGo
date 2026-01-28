# Makefile
PROJECT_NAME := SublimeGo

.PHONY: all build run test clean lint audit css

all: build

# Installation des outils (Air pour le hot reload, Templ pour les vues)
install-tools:
	go install github.com/air-verse/air@latest
	go install github.com/a-h/templ/cmd/templ@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Génération du code (Templ + Ent plus tard)
generate:
	templ generate
	# go generate ./internal/ent/... (on décommentera plus tard)

# Compilation CSS avec Tailwind (mode développement)
css:
	.\tailwindcss.exe -i .\pkg\ui\assets\input.css -o .\pkg\ui\assets\output.css --watch

# Lancer en mode développement (Hot Reload)
dev: generate
	air

# Construire le binaire final (Production)
build: generate
	go build -o bin/$(PROJECT_NAME) cmd/app/main.go

# Vérifier la qualité du code (Strict)
lint:
	golangci-lint run

# Nettoyer
clean:
	rm -rf bin
	rm -rf tmp