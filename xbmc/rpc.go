package xbmc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"runtime"
	"sync"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/pdf/xbmc-callback-daemon/logger"
	"github.com/stefantalpalaru/pool"
)

const MinVersion = 6

type Connection struct {
	conn          net.Conn
	write         chan interface{}
	Notifications chan Notification
	enc           *json.Encoder
	dec           *json.Decoder
	lock          sync.Mutex
	requestId     uint32
	responses     map[uint32]chan *Response
	pool          *pool.Pool

	address string
}

type Error struct {
	Code    float64                 `json:"code"`
	Message string                  `json:"message"`
	Data    *map[string]interface{} `json:"data"`
}

type Response struct {
	Id      *float64                `json:"id"`
	JsonRPC string                  `json:"jsonrpc"`
	Method  *string                 `json:"method"`
	Params  *map[string]interface{} `json:"params"`
	Result  *map[string]interface{} `json:"result"`
	Error   *Error                  `json:"error"`
}

type QueryLimits struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

type QueryParams struct {
	Properties []string     `json:"properties,omitempty"`
	Limits     *QueryLimits `json:"limits,omitempty"`
}

// Notification stores XBMC server->client notifications.
type Notification struct {
	Method string `json:"method" mapstructure:"method"`
	Params struct {
		Data struct {
			Item *struct {
				Type string `json:"type" mapstructure:"type"`
			} `json:"item" mapstructure:"item"` // Optional
		} `json:"data" mapstructure:"data"`
	} `json:"params" mapstructure:"params"`
}

// Unpack the response and errors from the combined object out of a channel
func (res *Response) Unpack() (result map[string]interface{}, err error) {
	if res.Error != nil {
		err = errors.New(fmt.Sprintf(
			`XBMC error (%v): %v`, res.Error.Code, res.Error.Message,
		))
	} else if res.Result != nil {
		result = *res.Result
	} else {
		logger.Debug(`Received unknown response type from XBMC: `, res)
	}
	return result, err
}

// New brings up an instance of the XBMC connection
func (c *Connection) New(address string) (err error) {
	if c.address == `` {
		c.address = address
	}

	if err = c.Connect(); err != nil {
		return err
	}

	c.write = make(chan interface{}, 4)
	c.Notifications = make(chan Notification, 4)

	c.responses = make(map[uint32]chan *Response)

	c.enc = json.NewEncoder(c.conn)
	c.dec = json.NewDecoder(c.conn)

	c.pool = pool.New(runtime.NumCPU() * 3)
	c.pool.Run()

	go c.reader()
	go c.writer()

	ch, err := c.Request(`JSONRPC.Version`, nil)
	if err != nil {
		return err
	}

	r := <-ch
	res, err := r.Unpack()
	if err != nil {
		return err
	}

	if version := res[`version`].(map[string]interface{}); version != nil {
		if version[`major`].(float64) < MinVersion {
			return errors.New(`XBMC version too low, upgrade to Frodo or later`)
		}
	}
	logger.Info(`Connected to XBMC`)

	return
}

func (c *Connection) Request(method string, params interface{}) (ch chan *Response, err error) {
	c.lock.Lock()
	id := c.requestId
	ch = make(chan *Response)
	c.responses[id] = ch
	c.requestId++
	c.lock.Unlock()

	var r map[string]interface{}

	// I'd really like to be using structs here, but we have to do this
	// shit because XBMC breaks with null params, and Go has fucked JSON
	// encoding by default:
	// https://code.google.com/p/go/issues/detail?id=5452
	if params == nil {
		r = map[string]interface{}{
			"id":      id,
			`method`:  method,
			`jsonrpc`: `2.0`,
		}
	} else {
		r = map[string]interface{}{
			"id":      id,
			`method`:  method,
			`params`:  params,
			`jsonrpc`: `2.0`,
		}
	}
	logger.Debug(`Sending XBMC Request: `, r)
	c.write <- r

	return ch, nil
}

func (c *Connection) Notification(method string, params interface{}) (err error) {
	var r map[string]interface{}

	// I'd really like to be using structs here, but we have to do this
	// shit because XBMC breaks with null params, and Go has fucked JSON
	// encoding by default:
	// https://code.google.com/p/go/issues/detail?id=5452
	if params == nil {
		r = map[string]interface{}{
			`method`:  method,
			`jsonrpc`: `2.0`,
		}
	} else {
		r = map[string]interface{}{
			`method`:  method,
			`params`:  params,
			`jsonrpc`: `2.0`,
		}
	}
	logger.Debug(`XBMC Notification: `, r)
	c.write <- r

	return
}

// Connect establishes a TCP connection
func (c *Connection) Connect() (err error) {
	c.conn, err = net.Dial(`tcp`, c.address)
	for err != nil {
		logger.Error(`Connecting to XBMC: `, err)
		logger.Info(`Attempting reconnect...`)
		time.Sleep(time.Second)
		c.conn, err = net.Dial(`tcp`, c.address)
	}
	err = nil

	return
}

// writer loop processes outbound requests
func (c *Connection) writer() {
	for {
		var req interface{}
		req = <-c.write
		if err := c.enc.Encode(req); err != nil {
			logger.Warn(`Failed encoding request for XBMC: `, err)
			break
		}
	}
}

// reader loop processes inbound responses and notifications
func (c *Connection) reader() {
	for {
		res := new(Response)
		err := c.dec.Decode(res)
		if err == io.EOF {
			logger.Error(`Reading from XBMC: `, err)
			logger.Error(`If this error persists, make sure you are using the JSON-RPC port, not the HTTP port!`)
			time.Sleep(time.Second)
			if err = c.Connect(); err != nil {
				logger.Error(`Reconnecting to XBMC: `, err)
			}
		} else if err != nil {
			logger.Error(`Decoding response from XBMC: `, err)
		} else {
			if res.Id == nil && res.Method != nil {
				logger.Debug(`Received notification from XBMC: `, *res.Method)
				n := Notification{}
				n.Method = *res.Method
				mapstructure.Decode(res.Params, &n.Params)
				c.Notifications <- n
			} else {
				if ch := c.responses[uint32(*res.Id)]; ch != nil {
					if res.Result != nil {
						logger.Debug(`Received response from XBMC: `, *res.Result)
					}
					ch <- res
				} else {
					logger.Warn(
						`Received XBMC response for unknown request: `,
						*res.Id,
					)
					logger.Debug(
						`Current response channels: `, c.responses,
					)
				}
			}
		}
	}
}

// Close XBMC connection
func (c *Connection) Close() {
	for _, v := range c.responses {
		close(v)
	}
	close(c.write)
	close(c.Notifications)
	c.pool.Stop()
	c.conn.Close()
}

// Execute takes the callback and performs a JSON-RPC request over the
// established XBMC connection.
func (c *Connection) Execute(callback map[string]interface{}) (err error) {
	logger.Debug(`Sending request to XBMC: `, callback)

	ch, err := c.Request(callback[`method`].(string), callback[`params`])
	if err != nil {
		logger.Error(`Could not send request to XBMC: `, err)
		return err
	}

	select {
	case r := <-ch:
		_, err = r.Unpack()
		if err != nil {
			logger.Warn(`XBMC responded: `, err)
			return err
		}
	case <-time.After(1 * time.Second):
		logger.Warn(`Timeout waiting for XBMC response`)
	}

	return
}
