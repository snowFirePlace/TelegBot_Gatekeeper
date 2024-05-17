package main

import (
	"fmt"
	"os"
	"time"

	"net/http"
	_ "net/http/pprof"

	"github.com/snowFirePlace/logger"
	"snowfireplace.com/config"
	"snowfireplace.com/data"
	"snowfireplace.com/network"
	"snowfireplace.com/telegram"
	// "github.com/SnowPlace/logger"
)

func main() {
	data.Directory = "./"
	// data.Directory, _ = filepath.Abs(filepath.Dir(os.Args[0]))
	logger.New(data.Directory + string(os.PathSeparator) + "bot.log")
	data.StartTime = time.Now()
	config.Get(data.Directory + string(os.PathSeparator) + "config.yml")
	addrs := config.Addrs
	t := 60
	bad := map[int]int{}
	funcMap := func() {
		time.Sleep(time.Second * 120)
		for {
			for j, n := range addrs {
				addr := n.Address
				if len(bad) == 0 && t == 1 {
					t = 60
				}
				if n.Type == "http" {
					if n.State {
						if t != 60 {
							time.Sleep(60 * time.Second)
						}
						if err := network.Response(addr); err != nil {
							n.State = false
							bad[len(bad)] = j
							addrs[j] = n
							// fmt.Println("Send message bad")
							t = 1
							data.MessageChan <- fmt.Sprintf("%s не работает", n.Name)
						}
					} else {
						if t != 60 {
							time.Sleep(30 * time.Second)
						}
						if err := network.Response(addr); err == nil {
							n.State = true
							addrs[j] = n
							delete(bad, j)
							// fmt.Println("Send message bad")
							data.MessageChan <- fmt.Sprintf("Работа %s восстановлена", n.Name)
						}
					}
				} else {
					if n.State {
						if err := network.Ping(addr, 3); err != nil {
							n.State = false
							bad[len(bad)] = j
							addrs[j] = n
							// fmt.Println("Send message bad")
							t = 1
							data.MessageChan <- fmt.Sprintf("Связь с %s потеряна", n.Name)
						}
					} else {
						if err := network.Ping(addr, 30); err == nil {
							n.State = true
							addrs[j] = n
							delete(bad, j)
							// fmt.Println("Send message good")
							data.MessageChan <- fmt.Sprintf("Связь с %s восстановлена", n.Name)
						}
					}
				}
			}
			time.Sleep(time.Duration(t) * time.Second)
		}
	}
	go funcMap()
	funcTeleg := func() { telegram.Run() }
	go funcTeleg()
	http.ListenAndServe("0.0.0.0:8080", nil)
}
