package data

import "time"

const (
	Version = "0.4.0_20.05.18"
)

var (
	MessageChan chan string = make(chan string)
	StartTime   time.Time
	Directory   string
)
