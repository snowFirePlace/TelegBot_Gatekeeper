.PHONY:
.SILENT:

build:
	go build -o ./.bin/bot -ldflags="-X 'main.version=$(git describe --abbrev=0)'" cmd/bot/main.go

run: build
	./.bin/bot