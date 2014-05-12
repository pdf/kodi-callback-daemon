package main

import (
	`fmt`
	`os`
	`github.com/pdf/xbmc-callback-daemon/config`
	`github.com/pdf/xbmc-callback-daemon/hyperion`
	`github.com/pdf/xbmc-callback-daemon/logger`
	`github.com/pdf/xbmc-callback-daemon/shell`
	`github.com/pdf/xbmc-callback-daemon/xbmc`
)

const (
	VERSION = `0.2.0`
)

var (
	cfg config.Config
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
	cfg = config.Load(os.Args[1])
	if cfg.Debug != nil {
		logger.DebugEnabled = *cfg.Debug
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
			xbmc.Execute(m)

		case `shell`:
			shell.Execute(m)

		default:
			logger.Warn(`Unknown backend: `, m[`backend`])
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
				logger.Panic(`Couldn't understand 'types' array, check your configuration.`)
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
	// Connect to XBMC, this is required.
	xbmc.Connect(fmt.Sprintf(`%s:%d`, cfg.XBMC.Address, cfg.XBMC.Port))
	defer xbmc.Close()

	// If the configuration specifies a Hyperion connection, use it.
	if cfg.Hyperion != nil {
		hyperion.Connect(fmt.Sprintf(`%s:%d`, cfg.Hyperion.Address, cfg.Hyperion.Port))
		defer hyperion.Close()
	}

	notification := &xbmc.Notification{}
	// Get callbacks from configuration.
	callbacks := cfg.Callbacks.(map[string]interface{})

	// Execute callbacks for the special `Startup` notification.
	if callbacks[`Startup`] != nil {
		execute(callbacks[`Startup`].([]interface{}))
	}

	// Loop while reading from XBMC.
	for {
		// Read from XBMC.
		xbmc.Read(notification)

		logger.Debug(`Received notification from XBMC: `, notification)
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
