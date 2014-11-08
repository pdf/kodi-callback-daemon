package xbmc

import (
	"github.com/StreamBoat/xbmc_jsonrpc"

	. "github.com/pdf/xbmc-callback-daemon/log"
)

// Execute takes an XBMC JSON-RPC Connection and a callback, and performs the
// RPC request contained in the callback
func Execute(x *xbmc_jsonrpc.Connection, callback map[string]interface{}) {
	Logger.Debug(`Sending request to XBMC: %v`, callback)

	req := xbmc_jsonrpc.Request{}
	req.Method = callback[`method`].(string)
	if callback[`params`] != nil {
		params := callback[`params`].(map[string]interface{})
		req.Params = &params
	}
	_ = x.Send(req, false)
}
