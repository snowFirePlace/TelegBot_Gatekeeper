package config

import (
	"os"

	yaml "gopkg.in/yaml.v2"
)

var (
	cfg     Config
	Version string
)

type Config struct {
	TelegramBotToken string `yaml:"telegram_bot_token"`
	TelegramChannel  int64  `yaml:"telegram_channel_id"`
	Admin            Admin
}

type Admin struct {
	ID     int    `yaml:"id"`
	FIO    string `yaml:"fio"`
	Branch string `yaml:"branch"`
	Unit   string `yaml:"unit"`
	Phone  string `yaml:"phone"`
}

func Get() Config {
	yamlFile, err := os.ReadFile("config.yml")
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(yamlFile, &cfg)
	if err != nil {
		panic(err)
	}
	return cfg
}
