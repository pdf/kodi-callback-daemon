package xbmc

import (
	"github.com/StreamBoat/xbmc_jsonrpc"

	log "github.com/Sirupsen/logrus"
)

// Execute takes an XBMC JSON-RPC Connection and a callback, and performs the
// RPC request contained in the callback
func Execute(x *xbmc_jsonrpc.Connection, callback map[string]interface{}) {
	log.WithField(`request`, callback).Debug(`Sending to XBMC`)

	req := xbmc_jsonrpc.Request{}
	req.Method = callback[`method`].(string)
	if callback[`params`] != nil {
		params := callback[`params`].(map[string]interface{})
		req.Params = &params
	}
	_ = x.Send(req, false)
}
