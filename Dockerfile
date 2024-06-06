FROM golang:1.22-alpine AS builder
WORKDIR /usr/local/src

# RUN apk --no-cache add bash git make gcc musl-dev
RUN apk --no-cache add gcc git g++


COPY ["go.mod", "go.sum","./"]

RUN go mod download 

#build
COPY .git ./.git
COPY cmd ./cmd
COPY internal ./internal
RUN go build -o ./bin/app -ldflags="-X 'main.version=$(git describe --tags)'" cmd/bot/main.go

FROM alpine:latest AS runner

COPY --from=builder /usr/local/src/bin/app /

VOLUME [ "/config" ]
VOLUME [ "/data" ]

CMD ["/app"]