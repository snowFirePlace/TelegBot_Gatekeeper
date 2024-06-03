.PHONY:
.SILENT:

build:
	go build -o ./.bin/bot -ldflags="-X 'main.version=$(git rev-parse --short HEAD)'" cmd/bot/main.go

run: build
	./.bin/bot