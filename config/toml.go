package config

import (
	"os"

	"github.com/BurntSushi/toml"
)

type arpSectionConfig struct {
	Log string
	Db  string
}

type zabbixSectionConfig struct {
	Sender string
	Config string
	Log    string
}

type Config struct {
	Arp    arpSectionConfig    `toml:"arp"`
	Zabbix zabbixSectionConfig `toml:"zabbix"`
}

func Parse(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	conf := new(Config)
	if _, err := toml.Decode(string(data[:]), conf); err != nil {
		return nil, err
	}
	return conf, nil
}
