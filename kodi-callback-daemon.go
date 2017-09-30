package main

import (
	"fmt"
	"os"
	"time"

	"github.com/pdf/golifx"
	"github.com/pdf/kodi-callback-daemon/boblight"
	"github.com/pdf/kodi-callback-daemon/config"
	"github.com/pdf/kodi-callback-daemon/hyperion"
	"github.com/pdf/kodi-callback-daemon/kodi"
	"github.com/pdf/kodi-callback-daemon/lifx"
	"github.com/pdf/kodi-callback-daemon/shell"
	"github.com/pdf/kodirpc"

	"github.com/Sirupsen/logrus"
)

const (
	// VERSION of the application
	VERSION = "1.5.0"
	// LIFXDELAY delays startup execution to allow LIFX devices to report their
	// group membership
	LIFXDELAY = 3 * time.Second
)

var (
	log *logrus.Logger
	cfg config.Config
	k   *kodirpc.Client
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
	log = logrus.New()

	// Initialize logger
	cfg = config.Load(os.Args[1])
	if cfg.Debug != nil && *cfg.Debug {
		log.Level = logrus.DebugLevel
	} else {
		log.Level = logrus.InfoLevel
	}
}

type handler struct {
	params map[string]interface{}
}

func (h *handler) typedCallback(method string, data interface{}) {
	switch h.params[`types`].(type) {
	case []interface{}:
		d, ok := data.(map[string]interface{})
		if !ok {
			log.WithField(`data`, data).Error(`Invalid notification structure`)
			return
		}
		for _, t := range h.params[`types`].([]interface{}) {
			item, ok := d[`item`].(map[string]interface{})
			if !ok {
				log.WithField(`data`, data).Error(`Invalid notification structure`)
				return
			}
			if t.(string) == item[`type`] {
				h.callback(method, data)
			}
		}
	default:
		h.callback(method, data)
	}
}

func (h *handler) callback(method string, data interface{}) {
	switch h.params[`backend`].(string) {
	case `hyperion`:
		if cfg.Hyperion != nil {
			hyperion.Execute(h.params)
		}

	case `lifx`:
		if cfg.LIFX != nil {
			lifx.Execute(h.params)
		}

	case `kodi`, `xbmc`:
		kodi.Execute(k, h.params)

	case `shell`:
		shell.Execute(h.params)

	default:
		log.WithField(`backend`, h.params[`backend`]).Warn(`Unknown backend`)
	}
}

// main program loop.
func main() {
	var host config.Host
	// Connect to Kodi, this is required.
	kodirpc.SetLogger(log)
	kodiConfig := kodirpc.NewConfig()
	kodiConfig.ConnectTimeout = 0
	if cfg.Kodi != nil {
		if cfg.Kodi.Timeout != nil {
			kodiConfig.ReadTimeout = time.Duration(*cfg.Kodi.Timeout)
		}
		host = *cfg.Kodi
	} else if cfg.XBMC != nil {
		if cfg.XBMC.Timeout != nil {
			kodiConfig.ReadTimeout = time.Duration(*cfg.XBMC.Timeout)
		}
		host = *cfg.XBMC
	} else {
		log.Fatal(`You must provide Kodi/XBMC connection details in your configuration`)
	}

	k, err := kodirpc.NewClient(
		fmt.Sprintf(`%s:%d`, host.Address, host.Port),
		kodiConfig,
	)
	defer func() {
		if err = k.Close(); err != nil {
			log.Panicln(err)
		}
	}()

	if err != nil {
		log.WithField(`error`, err).Fatal(`Failed to obtain Kodi connection`)
	}

	// If the configuration specifies a Hyperion connection, use it.
	if cfg.Hyperion != nil {
		hyperion.Connect(&cfg)
		defer hyperion.Close()
	}

	// If the configuration specifies a LIFX connection, use it.
	if cfg.LIFX != nil {
		golifx.SetLogger(log)

		lifx.Connect(&cfg)
		defer lifx.Close()
	}

	// If the configuration specifies a boblight connection, use it.
	if cfg.Boblight != nil {
		boblight.Connect(&cfg)
		defer boblight.Close()
	}

	// Get callbacks from configuration.
	callbacks := cfg.Callbacks.(map[string]interface{})

	var h *handler
	// Execute callbacks for the special `Startup` notification.
	if callbacks[`Startup`] != nil {
		// LIFX groups take some time to populate, so we add an arbitrary delay
		// here to try and compensate
		if cfg.LIFX != nil {
			time.Sleep(LIFXDELAY)
		}

		for _, cb := range callbacks[`Startup`].([]interface{}) {
			h = &handler{params: cb.(map[string]interface{})}
			h.callback(``, nil)
		}
	}

	for key, value := range callbacks {
		for _, cb := range value.([]interface{}) {
			h = &handler{}
			h.params = cb.(map[string]interface{})
			switch key {
			case `Player.OnPlay`, `Player.OnPause`, `Player.OnStop`:
				k.Handle(key, h.typedCallback)
			default:
				k.Handle(key, h.callback)
			}
		}
	}

	wait := make(chan struct{})
	defer close(wait)
	<-wait
}
