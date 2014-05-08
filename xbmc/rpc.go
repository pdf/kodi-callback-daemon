package xbmc

import (
	`encoding/json`
	`io`
	`log`
	`net`
)

var conn net.Conn
var Decoder *json.Decoder

// XBMC server->client notification
type XBMCNotification struct {
	Method string
	Params struct {
		Data struct {
			Item *struct {
				Type string
			}
		}
	}
}

// Initialize connection
func Connect(address string) {
	conn, err := net.Dial(`tcp`, address)
	if err != nil {
		log.Panicln(`[ERROR] Connecting to XBMC: `, err)
	} else {
		log.Println(`[INFO] Connected to XBMC`)
	}
	Decoder = json.NewDecoder(conn)
}

// Send encoded data to Hyperion
func Read(notification *XBMCNotification) {
	err := Decoder.Decode(&notification)
	if err == io.EOF {
		log.Panicln(`[ERROR] Reading from XBMC: `, err)
	} else if err != nil {
		log.Println(`[ERROR] Decoding response from XBMC: `, err)
	}
}

func Close() {
	conn.Close()
}
