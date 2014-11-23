package config

import (
	"encoding/json"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
)

// Host contains the IP address, port and timeout for TCP/UDP connections.
type Host struct {
	Address string         `json:"address"` // Required
	Port    uint16         `json:"port"`    // Required
	Timeout *time.Duration `json:"timeout"` // Optional
}

// Config stores the json configuration structure.
type Config struct {
	XBMC      *Host       `json:"xbmc"`      // Deprecated
	Kodi      *Host       `json:"kodi"`      // Required (if XBMC not provided)
	Hyperion  *Host       `json:"hyperion"`  // Optional
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
