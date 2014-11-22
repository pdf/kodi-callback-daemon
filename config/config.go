package config

import (
	"encoding/json"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
)

// address contains the IP address and port for TCP/UDP connections.
type address struct {
	Address string         `json:"address"`
	Port    uint16         `json:"port"`
	Timeout *time.Duration `json:"timeout"`
}

// Config stores the json configuration structure.
type Config struct {
	XBMC      address     `json:"xbmc"`      // Required
	Hyperion  *address    `json:"hyperion"`  // Optional
	Debug     *bool       `json:"debug"`     // Optional
	Callbacks interface{} `json:"callbacks"` // Required
}

// Load reads the configuration from the specified filename, and returns the
// decoded JSON data.
func Load(filename string) Config {
	file, err := os.Open(filename)
	if err != nil {
		log.WithField(`error`, err).Fatal(`Opening config file`)
	}

	dec := json.NewDecoder(file)
	conf := Config{}
	if err = dec.Decode(&conf); err != nil {
		log.WithField(`error`, err).Fatal(`Parsing config file`)
	}

	return conf
}
