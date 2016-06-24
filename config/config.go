package config

import (
	"encoding/json"
	"errors"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
)

// Timeout is a duration type that unmarshals from a JSON string
type Timeout time.Duration

// UnmarshalJSON converts a time string into a time.Duration
func (t *Timeout) UnmarshalJSON(b []byte) error {
	var (
		timeout time.Duration
		str     string
		err     error
	)

	if err = json.Unmarshal(b, &str); err != nil {
		return errors.New(`Error interpreting timeout as string`)
	}

	if timeout, err = time.ParseDuration(str); err != nil {
		return errors.New(`Error interpreting timeout as duration`)
	}
	converted := Timeout(timeout)
	t = &converted

	return nil
}

// Host contains the IP address, port and timeout for TCP/UDP connections.
type Host struct {
	Address string   `json:"address"` // Required
	Port    uint16   `json:"port"`    // Required
	Timeout *Timeout `json:"timeout"` // Optional
}

// LIFX stores the LIFX json configuration structure.
type LIFX struct {
	Timeout *Timeout `json:"timeout"` // Optional
}

// Boblight stores the boblight json configuration structure
type Boblight struct {
	Input  *Host `json:"input"`  // Required
	Output *Host `json:"output"` // Required
}

// Hyperion stores the hyperion json configuration structure
type Hyperion struct {
	Input  *Host `json:"input"`  // Optional
	Output *Host `json:"output"` // Required, unless legacy configuration
	*Host        // Embedded host to support legacy configuration
}

// Config stores the json configuration structure.
type Config struct {
	XBMC      *Host       `json:"xbmc"`      // Deprecated
	Kodi      *Host       `json:"kodi"`      // Required (if XBMC not provided)
	Hyperion  *Hyperion   `json:"hyperion"`  // Optional
	LIFX      *LIFX       `json:"lifx"`      // Optional
	Boblight  *Boblight   `json:"boblight"`  // Optional
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
