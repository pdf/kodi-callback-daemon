package xbmc

import (
	`encoding/json`
	`io`
	`net`
	`github.com/pdf/xbmc-callback-daemon/logger`
)

var conn net.Conn
var decoder *json.Decoder

// Notification stores XBMC server->client notifications.
type Notification struct {
	Method string `json:"method"`
	Params struct {
		Data struct {
			Item *struct {
				Type string `json:"type"`
			} `json:"item"` // Optional
		} `json:"data"`
	} `json:"params"`
}

// Connect establishes a TCP connection to the specified address and attaches
// JSON decoder.
func Connect(address string) {
	conn, err := net.Dial(`tcp`, address)
	if err != nil {
		logger.Panic(`Connecting to XBMC: `, err)
	} else {
		logger.Info(`Connected to XBMC`)
	}
	decoder = json.NewDecoder(conn)
}

// Read and decode JSON from the XBMC connection into the notification pointer.
func Read(notification *Notification) {
	err := decoder.Decode(&notification)
	// Bail on EOF, eat any decoding errors otherwise.
	// TODO: This probably needs to be more robust.
	if err == io.EOF {
		logger.Panic(`Reading from XBMC: `, err)
	} else if err != nil {
		logger.Error(`Decoding response from XBMC: `, err)
		return
	}
}

// Close XBMC connection
func Close() {
	conn.Close()
}

// Execute takes the callback and performs a JSON-RPC request over the
// established XBMC connection.
func Execute(callback map[string]interface{}) {
	logger.Debug(`Sending request to XBMC: `, callback)
	// BUG(pdf): xbmc.Execute is not implemented yet.
	logger.Warn(`xbmc.Execute(): Not implemented`)
}
