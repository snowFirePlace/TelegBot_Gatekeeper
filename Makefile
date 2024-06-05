
VERSION = $(shell git describe --abbrev=0)
.PHONY:
.SILENT:

build:
	echo "VERSION: $(VERSION)"
	docker build -t gatekeeper:$(VERSION) .
	# go build -o gatekeeper.exe -ldflags="-X 'main.version=$(git describe --abbrev=0)'" cmd/bot/main.go
	# go build -o ./.bin/bot -ldflags="-X 'main.version=$(git describe --abbrev=0)'" cmd/bot/main.go

run: build
	./.bin/bot