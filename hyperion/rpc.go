package hyperion

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"reflect"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/pdf/kodi-callback-daemon/config"
)

var (
	cfg     *config.Config
	conn    net.Conn
	encoder *json.Encoder
	decoder *json.Decoder
)

// Response stores Hyperion results for RPC calls.
type Response struct {
	Success bool    `json:"success"`
	Error   *string `json:"error"`
}

// qtfloat64 is a float64 class to ensure marshalling as floats as QT expects
// them.
// Seriously, fuck you QT.
type qtfloat64 float64

// Custom marshaller for correct float output
func (f qtfloat64) MarshalText() ([]byte, error) {
	v := reflect.ValueOf(f)
	return []byte(strconv.FormatFloat(v.Float(), 'f', 6, 64)), nil
}

// Custom marshaller for correct float output
func (f qtfloat64) MarshalJSON() ([]byte, error) {
	v := reflect.ValueOf(f)
	return []byte(strconv.FormatFloat(v.Float(), 'f', 6, 64)), nil
}

// Connect establishes a TCP connection to the specified address and attaches
// JSON encoders/decoders.
func Connect(conf *config.Config) {
	var (
		err        error
		outputAddr string
	)
	if cfg == nil {
		cfg = conf
	}
	if cfg.Hyperion.Output != nil {
		outputAddr = fmt.Sprintf(`%s:%d`, cfg.Hyperion.Output.Address, cfg.Hyperion.Output.Port)
	} else if cfg.Hyperion.Address != `` {
		outputAddr = fmt.Sprintf(`%s:%d`, cfg.Hyperion.Address, cfg.Hyperion.Port)
	} else {
		log.Error(`Missing Hyperion output configuration`)
		return
	}
	conn, err = net.Dial(`tcp`, outputAddr)
	for err != nil {
		log.WithField(`error`, err).Error(`Connecting to Hyperion`)
		log.Info(`Attempting reconnect...`)
		time.Sleep(time.Second)
		conn, err = net.Dial(`tcp`, outputAddr)
	}
	log.Info(`Connected to Hyperion`)
	encoder = json.NewEncoder(conn)
	decoder = json.NewDecoder(conn)

	if cfg.Hyperion.Input != nil {
		go func() {
			if err = listen(fmt.Sprintf(`%s:%d`, cfg.Hyperion.Input.Address, cfg.Hyperion.Input.Port)); err != nil {
				log.WithField(`error`, err).Error(`Hyperion UDP listener failed`)
			}
		}()
	}
}

// Close Hyperion connection
func Close() {
	_ = conn.Close()
}

// coerce takes a key/value pair and recurses down the value, replacing any
// float64 values with qtfloat64 conversions and returns the result.
// Some known non-float values are instead converted to integers.
// Seriously, fuck you QT.
func coerce(key string, value interface{}) interface{} {
	switch value.(type) {
	case float64:
		switch key {
		case `priority`:
			return uint16(value.(float64))
		case `color`:
			return uint8(value.(float64))
		default:
			return qtfloat64(value.(float64))
		}
	case []interface{}:
		result, ok := value.([]interface{})
		if !ok {
			log.WithField(`value`, value).Fatal(`Could not parse array, check configuration near`)
		}
		for i := range result {
			result[i] = coerce(key, result[i])
		}
		return result
	case map[string]interface{}:
		result, ok := value.(map[string]interface{})
		if !ok {
			log.WithField(`value`, value).Fatal(`Could not parse object, check configuration near`)
		}
		for k, v := range result {
			result[k] = coerce(k, v)
		}
		return result
	default:
		return value
	}
}

// Read and decode JSON from the Hyperion connection into the notification pointer.
func Read(response *Response) {
	err := decoder.Decode(&response)
	// Kick off the connection again on EOF, eat any decoding errors otherwise.
	// TODO: This probably needs to be more robust.
	if _, ok := err.(net.Error); err == io.EOF || ok {
		log.WithField(`error`, err).Error(`Reading from Hyperion`)
		Connect(cfg)
	} else if err != nil {
		log.WithField(`error`, err).Error(`Decoding response from Hyperion`)
		return
	}
}

// Execute takes the callback and performs a JSON-RPC request over the
// established Hyperion connection.
func Execute(callback map[string]interface{}) {
	response := &Response{}

	// Drop properties that the backend doesn't understand, and coerce float64/int
	cb := make(map[string]interface{}, 0)
	for k, v := range callback {
		switch k {
		case `backend`, `types`:
			continue
		default:
			cb[k] = coerce(k, v)
		}
	}

	log.WithField(`request`, cb).Debug(`Sending to Hyperion`)
	// Encode and send request
	err := encoder.Encode(&cb)
	if _, ok := err.(net.Error); ok {
		log.WithField(`error`, err).Error(`Writing to Hyperion`)
		Connect(cfg)
		if err = encoder.Encode(&cb); err != nil {
			log.WithField(`error`, err).Error(`Failed writing to Hyperion`)
		}
	} else if err != nil {
		log.WithField(`error`, err).Error(`Writing to Hyperion`)
	}
	// Check response and log any failure responses from Hyperion
	Read(response)
	log.WithField(`response`, response).Debug(`Received from Hyperion`)
	if !response.Success && response.Error != nil {
		log.WithField(`response.Error`, *response.Error).Warn(`Received from Hyperion`)
	}
}
