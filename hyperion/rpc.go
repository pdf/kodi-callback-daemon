package hyperion

import (
	"encoding/json"
	"io"
	"net"
	"reflect"
	"strconv"
	"time"

	. "github.com/pdf/xbmc-callback-daemon/log"
)

var (
	connAddress string
	conn        net.Conn
	encoder     *json.Encoder
	decoder     *json.Decoder
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
func Connect(address string) {
	if connAddress == `` {
		connAddress = address
	}
	conn, err := net.Dial(`tcp`, address)
	for err != nil {
		Logger.Error(`Connecting to Hyperion: %v`, err)
		Logger.Info(`Attempting reconnect...`)
		time.Sleep(time.Second)
		conn, err = net.Dial(`tcp`, address)
	}
	Logger.Info(`Connected to Hyperion`)
	encoder = json.NewEncoder(conn)
	decoder = json.NewDecoder(conn)
}

// Close Hyperion connection
func Close() {
	conn.Close()
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
		if ok == false {
			Logger.Fatalf(`Could not parse array, check configuration near %v`, value)
		}
		for i := range result {
			result[i] = coerce(key, result[i])
		}
		return result
	case map[string]interface{}:
		result, ok := value.(map[string]interface{})
		if ok == false {
			Logger.Fatalf(`Could not parse object, check configuration near %v`, value)
		}
		for k, v := range result {
			result[k] = coerce(k, v)
		}
		return result
	default:
		return value
	}
}

// Read and decode JSON from the XBMC connection into the notification pointer.
func Read(response *Response) {
	err := decoder.Decode(&response)
	// Kick off the connection again on EOF, eat any decoding errors otherwise.
	// TODO: This probably needs to be more robust.
	if _, ok := err.(net.Error); err == io.EOF || ok {
		Logger.Error(`Reading from Hyperion: %v`, err)
		Connect(connAddress)
	} else if err != nil {
		Logger.Error(`Decoding response from Hyperion: %v`, err)
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

	Logger.Debug(`Sending request to Hyperion: %v`, cb)
	// Encode and send request
	err := encoder.Encode(&cb)
	if _, ok := err.(net.Error); ok {
		Logger.Error(`Writing to Hyperion: %v`, err)
		Connect(connAddress)
		encoder.Encode(&cb)
	} else if err != nil {
		Logger.Error(`Sending to Hyperion: %v`, err)
	}
	// Check response and log any failure responses from Hyperion
	Read(response)
	Logger.Debug(`Received response from Hyperion: `, response)
	if response.Success == false && response.Error != nil {
		Logger.Warning(`Error received from Hyperion: `, *response.Error)
	}
}
