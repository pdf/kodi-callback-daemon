package main

import (
	"fmt"
	"os"
	"time"

	"github.com/StreamBoat/kodi_jsonrpc"
	"github.com/pdf/kodi-callback-daemon/config"
	"github.com/pdf/kodi-callback-daemon/hyperion"
	"github.com/pdf/kodi-callback-daemon/kodi"
	"github.com/pdf/kodi-callback-daemon/lifx"
	"github.com/pdf/kodi-callback-daemon/shell"

	log "github.com/Sirupsen/logrus"
)

const (
	// VERSION of the application
	VERSION = "1.1.5"
)

var (
	cfg config.Config
	k   kodi_jsonrpc.Connection
)

// usage simply prints the invocation requirements.
func usage() {
	fmt.Fprintf(os.Stderr, "\nKodi Callback Daemon v%s\n\nUsage: %s [configFile]\n\n", VERSION, os.Args[0])
	os.Exit(1)
}

// init ensures we have a config path argument, and loads the configuration.
func init() {
	if len(os.Args) < 2 {
		usage()
	}

	// Initialize logger
	cfg = config.Load(os.Args[1])
	if cfg.Debug != nil && *cfg.Debug == true {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
}

// execute iterates through a list of callbacks, and sends them to the backend
// defined in the `backend` property.
func execute(callbacks []interface{}) {
	for i := range callbacks {
		m := callbacks[i].(map[string]interface{})

		switch m[`backend`] {
		case `hyperion`:
			if cfg.Hyperion != nil {
				hyperion.Execute(m)
			}

		case `lifx`:
			if cfg.LIFX != nil {
				lifx.Execute(m)
			}

		case `kodi`, `xbmc`:
			kodi.Execute(&k, m)

		case `shell`:
			shell.Execute(m)

		default:
			log.WithField(`backend`, m[`backend`]).Warn(`Unknown backend`)
		}
	}
}

// callbacksByType takes a type to match, and a list of callbacks. If the
// callback has a `types` property, and that contains a matching type, the
// callback is added to the returned list.  A callback without a `types`
// property will always be returned.
func callbacksByType(matchType string, callbacks []interface{}) []interface{} {
	var result []interface{}
	var cb map[string]interface{}

	for i := range callbacks {
		// Access internal callback map.
		cb = callbacks[i].(map[string]interface{})

		switch cb[`types`].(type) {
		// We have a list of types.
		case []interface{}:
			// Access internal types slice.
			cbTypes, ok := cb[`types`].([]interface{})
			if ok == false {
				log.Fatal(`Couldn't understand 'types' array, check your configuration.`)
			}
			for j := range cbTypes {
				if cbTypes[j].(string) == matchType {
					// Matched the required type, add this callback to the results.
					result = append(result, cb)
				}
			}
		// If there is no valid `types` property, add this callback to the results.
		default:
			result = append(result, cb)
		}
	}

	return result
}

// main program loop.
func main() {
	var host config.Host
	// Connect to Kodi, this is required.
	kodiTimeout := time.Duration(0)
	if cfg.Kodi != nil {
		if cfg.Kodi.Timeout != nil {
			kodiTimeout = time.Duration(*cfg.Kodi.Timeout)
		}
		host = *cfg.Kodi
	} else if cfg.XBMC != nil {
		if cfg.XBMC.Timeout != nil {
			kodiTimeout = time.Duration(*cfg.XBMC.Timeout)
		}
		host = *cfg.XBMC
	} else {
		log.Fatal(`You must provide Kodi/XBMC connection details in your configuration`)
	}
	k, err := kodi_jsonrpc.New(
		fmt.Sprintf(`%s:%d`, host.Address, host.Port),
		kodiTimeout,
	)

	defer k.Close()
	if err != nil {
		log.WithField(`error`, err).Fatal(`Failed to obtain Kodi connection`)
	}

	// If the configuration specifies a Hyperion connection, use it.
	if cfg.Hyperion != nil {
		hyperion.Connect(fmt.Sprintf(`%s:%d`, cfg.Hyperion.Address, cfg.Hyperion.Port))
		defer hyperion.Close()
	}

	// If the configuration specifies a LIFX connection, use it.
	if cfg.LIFX != nil {
		lifx.Connect(cfg)
		defer lifx.Close()
	}

	// Get callbacks from configuration.
	callbacks := cfg.Callbacks.(map[string]interface{})

	// Execute callbacks for the special `Startup` notification.
	if callbacks[`Startup`] != nil {
		execute(callbacks[`Startup`].([]interface{}))
	}

	// Loop while reading from Kodi.
	for {
		// Read from Kodi.
		notification := <-k.Notifications

		// Match Kodi notification to our configured callbacks.
		if callbacks[notification.Method] != nil {
			cbs := callbacks[notification.Method].([]interface{})
			switch notification.Method {
			case `Player.OnPlay`, `Player.OnPause`, `Player.OnStop`:
				cbs = callbacksByType(notification.Params.Data.Item.Type, cbs)
			}
			execute(cbs)
		}
	}
}
