package proxy

import (
	"io/ioutil"
	"log"
	"strings"
)

func GetList() (list []string) {
	dat, err := ioutil.ReadFile("././proxy.txt")
	if err != nil {
		log.Panic(err)
	}
	// resp, _ := http.Get("https://raw.githubusercontent.com/TheSpeedX/SOCKS-List/master/socks5.txt")
	// bytes, _ := ioutil.ReadAll(resp.Body)
	// for _, l := range strings.Split(string(bytes), "\n") {
	// key := false
	for _, l := range strings.Split(string(dat), "\n") {
		// if key {
		// arg := strings.Split(strings.Split(l, " ")[1], "-")
		// if arg[0] == "RU" {
		// 	continue
		// }
		// if arg[1] != "H" {
		// 	continue
		// }
		// address := strings.Split(l, " ")[0]
		// if strings.Split(l, ":")[1] == "8080" {
		list = append(list, l[:len(l)-1])
		// }
		// }
		// if !key && len(l) == 0 {
		// 	key = true
		// }
	}
	return
}
