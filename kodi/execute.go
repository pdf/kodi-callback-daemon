package kodi

import (
	"github.com/StreamBoat/kodi_jsonrpc"

	log "github.com/Sirupsen/logrus"
)

// Execute takes an Kodi/XBMC JSON-RPC Connection and a callback, and performs
// the RPC request contained in the callback
func Execute(k *kodi_jsonrpc.Connection, callback map[string]interface{}) {
	log.WithField(`request`, callback).Debug(`Sending to Kodi`)

	req := kodi_jsonrpc.Request{}
	req.Method = callback[`method`].(string)
	if callback[`params`] != nil {
		params := callback[`params`].(map[string]interface{})
		req.Params = &params
	}
	_ = k.Send(req, false)
}
