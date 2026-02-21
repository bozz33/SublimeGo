# Makefile - SublimeGo
# Multi-plateforme: Windows, Linux, macOS
# Sans dependance Node.js ou npm

PROJECT_NAME := SublimeGo

.PHONY: all build build-all run test clean lint css css-watch tailwind-download generate dev

all: build

# Installation des outils Go
install-tools:
	go install github.com/air-verse/air@latest
	go install github.com/a-h/templ/cmd/templ@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Telecharger Tailwind CSS Standalone CLI (une seule fois)
tailwind-download:
	go run ./cmd/tools/tailwind download

# Compilation CSS avec Tailwind v4 (production - minifie)
css:
	.\tailwindcss.exe -i ui\assets\css\input.css -o ui\assets\css\output.css --minify

# Compilation CSS en mode watch (developpement)
css-watch:
	.\tailwindcss.exe -i ui\assets\css\input.css -o ui\assets\css\output.css --watch

# Generation du code (Templ)
generate:
	templ generate

# Lancer en mode developpement (Hot Reload)
dev: generate
	air

# Construire le binaire final (Production)
build: generate css
	go build -o bin/$(PROJECT_NAME) ./cmd/sublimego

# Construire pour toutes les plateformes (cross-compilation pure Go, CGO_ENABLED=0)
build-all: generate css
	CGO_ENABLED=0 GOOS=linux   GOARCH=amd64  go build -ldflags="-s -w" -o bin/$(PROJECT_NAME)-linux-amd64   ./cmd/sublimego
	CGO_ENABLED=0 GOOS=linux   GOARCH=arm64  go build -ldflags="-s -w" -o bin/$(PROJECT_NAME)-linux-arm64   ./cmd/sublimego
	CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64  go build -ldflags="-s -w" -o bin/$(PROJECT_NAME)-darwin-amd64  ./cmd/sublimego
	CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64  go build -ldflags="-s -w" -o bin/$(PROJECT_NAME)-darwin-arm64  ./cmd/sublimego
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64  go build -ldflags="-s -w" -o bin/$(PROJECT_NAME)-windows-amd64.exe ./cmd/sublimego

# Verifier la qualite du code
lint:
	golangci-lint run

# Nettoyer
clean:
	rm -rf bin
	rm -rf tmp
	rm -rf tools