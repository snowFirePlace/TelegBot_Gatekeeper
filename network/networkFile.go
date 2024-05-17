package network

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	p "github.com/sparrc/go-ping"
)

func Ping(address string, t int) (err error) {
	pinger, err := p.NewPinger(address)
	if err != nil {
		return err
	}
	pinger.Count = 3
	pinger.Timeout = time.Duration(t) * time.Second
	pinger.SetPrivileged(true)
	pinger.Run() // blocks until finished
	stats := pinger.Statistics()
	if stats.PacketsRecv == 0 {
		return errors.New("Lost connect")
	}
	return
}
func Response(address string) (err error) {
	resp, err := http.Get(address)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf(resp.Status)
	}
	return

}
