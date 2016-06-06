package kodi

import (
	"github.com/pdf/kodirpc"

	log "github.com/Sirupsen/logrus"
)

// Execute takes an Kodi/XBMC JSON-RPC Connection and a callback, and performs
// the RPC request contained in the callback
func Execute(k *kodirpc.Client, callback map[string]interface{}) {
	log.WithField(`request`, callback).Debug(`Sending to Kodi`)

	var params map[string]interface{}
	if callback[`params`] != nil {
		params = callback[`params`].(map[string]interface{})
	}

	if err := k.Notify(callback[`method`].(string), params); err != nil {
		log.WithField(`request`, callback)
	}
}
