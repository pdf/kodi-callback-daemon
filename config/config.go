package config

import (
	`encoding/json`
	`log`
	`os`
)

// address contains the IP address and port for TCP/UDP connections.
type address struct {
	Address string `json:"address"`
	Port    uint16 `json:"port"`
}

// Config stores the json configuration structure.
type Config struct {
	XBMC      address     `json:"xbmc"`      // Required
	Hyperion  *address    `json:"hyperion"`  // Optional
	Callbacks interface{} `json:"callbacks"` // Required
}

// Load reads the configuration from the specified filename, and returns the
// decoded JSON data.
func Load(filename string) Config {
	file, err := os.Open(filename)
	if err != nil {
		log.Panicln(`[ERROR] Opening config file:`, err)
	}

	dec := json.NewDecoder(file)
	conf := Config{}
	if err = dec.Decode(&conf); err != nil {
		log.Panicln(`[ERROR] Parsing config file:`, err)
	}

	return conf
}
