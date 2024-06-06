VERSION = $(shell git describe --tags)

.PHONY:
.SILENT:

build:
	docker build -t gatekeeper:$(VERSION) .

run: build
	./.bin/bot
	
all: build