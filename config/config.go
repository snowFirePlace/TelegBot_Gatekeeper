package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

var (
	IDGroup int64
	Addrs   = map[int]object{}
	Proxy   bool
)

type object struct {
	Type    string `yaml:"type"`
	Address string `yaml:"address"`
	Name    string `yaml:"name"`
	State   bool
}
type db struct {
	IDGroup int64          `yaml:"idGroup"`
	Proxy   bool           `yaml:"proxy"`
	Objects map[int]object `yaml:"objects"`
}

func Get(path string) {
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		if err = Create(path); err != nil {
			log.Panic(fmt.Sprintf("Ошибка создания файла параметров."))
		}

	}
	DB := &db{}
	err = yaml.Unmarshal(yamlFile, DB)
	if err != nil {
		//e002
		log.Panic(fmt.Sprintf("|ERROR| Error reading the file\r\n"))
		log.Panic(fmt.Sprintf("Unmarshal: %v\r\n", err))
	}
	Addrs = DB.Objects
	for j, n := range Addrs {
		n.State = true
		Addrs[j] = n
	}

	if DB.Proxy {
		Proxy = true
	}
	IDGroup = DB.IDGroup
}
func Create(path string) (err error) {
	DB := &db{}
	DB.Objects = make(map[int]object)
	DB.Objects[1] = object{Type: "echo", Address: "127.0.0.1", Name: "localhost"}

	DB.Proxy = false
	a, _ := yaml.Marshal(&DB)
	t := strings.Split(string(a), "\n")
	text := []string{"---"}
	text = append(text, t...)
	ymlText := strings.Join(text, "\n")
	if _, err := os.Stat(path); err != nil {
		_ = ioutil.WriteFile(path, []byte(ymlText), 0644)
	}
	return nil
}
