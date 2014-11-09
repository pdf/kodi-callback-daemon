package main

import (
	"fmt"
	"os"
	"time"

	"github.com/StreamBoat/xbmc_jsonrpc"
	"github.com/pdf/xbmc-callback-daemon/config"
	"github.com/pdf/xbmc-callback-daemon/hyperion"
	"github.com/pdf/xbmc-callback-daemon/shell"
	"github.com/pdf/xbmc-callback-daemon/xbmc"

	. "github.com/pdf/xbmc-callback-daemon/log"
)

const (
	VERSION = "0.4.1"
)

var (
	cfg config.Config
	x   xbmc_jsonrpc.Connection
)

// usage simply prints the invocation requirements.
func usage() {
	fmt.Fprintf(os.Stderr, "\nXBMC Callback Daemon v%s\n\nUsage: %s [configFile]\n\n", VERSION, os.Args[0])
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
		SetLogLevel(`DEBUG`)
		xbmc_jsonrpc.SetLogLevel(`DEBUG`)
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

		case `xbmc`:
			xbmc.Execute(&x, m)

		case `shell`:
			shell.Execute(m)

		default:
			Logger.Warning(`Unknown backend: %v`, m[`backend`])
		}
	}
}

// callbacksByType takes a type to match, and a list of callbacks. If the
// callback has a `types` property, and that contains a matching type, the
// callback is added to the returned list.  A callback without a `types`
// property will always be returned.
func callbacksByType(matchType string, callbacks []interface{}) []interface{} {
	result := make([]interface{}, 0)
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
				Logger.Fatal(`Couldn't understand 'types' array, check your configuration.`)
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
	// Set XBMC client log level
	if cfg.Debug != nil && *cfg.Debug == true {
		xbmc_jsonrpc.SetLogLevel(`debug`)
	}
	// Connect to XBMC, this is required.
	xbmc_timeout := time.Duration(0)
	if cfg.XBMC.Timeout != nil {
		xbmc_timeout = *cfg.XBMC.Timeout
	}
	x, err := xbmc_jsonrpc.New(
		fmt.Sprintf(`%s:%d`, cfg.XBMC.Address, cfg.XBMC.Port),
		xbmc_timeout,
	)

	defer x.Close()
	if err != nil {
		Logger.Fatalf(`Failed to obtain XBMC connection: %v`, err)
	}

	// If the configuration specifies a Hyperion connection, use it.
	if cfg.Hyperion != nil {
		hyperion.Connect(fmt.Sprintf(`%s:%d`, cfg.Hyperion.Address, cfg.Hyperion.Port))
		defer hyperion.Close()
	}

	// Get callbacks from configuration.
	callbacks := cfg.Callbacks.(map[string]interface{})

	// Execute callbacks for the special `Startup` notification.
	if callbacks[`Startup`] != nil {
		execute(callbacks[`Startup`].([]interface{}))
	}

	// Loop while reading from XBMC.
	for {
		// Read from XBMC.
		notification := <-x.Notifications

		// Match XBMC notification to our configured callbacks.
		if callbacks[notification.Method] != nil {
			cbs := callbacks[notification.Method].([]interface{})
			// The Player.OnPlay notification supports an filtering by item type.
			if notification.Method == `Player.OnPlay` {
				cbs = callbacksByType(notification.Params.Data.Item.Type, cbs)
			}
			execute(cbs)
		}
	}
}
