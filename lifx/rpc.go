package lifx

import (
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/mitchellh/mapstructure"

	"github.com/pdf/golifx"
	lifxcommon "github.com/pdf/golifx/common"
	"github.com/pdf/golifx/protocol"
	"github.com/pdf/kodi-callback-daemon/boblight"
	"github.com/pdf/kodi-callback-daemon/common"
	"github.com/pdf/kodi-callback-daemon/config"
	"github.com/pdf/kodi-callback-daemon/hyperion"
)

type closer chan struct{}

type ledCallbackType uint8

const (
	ledCallbackBoblight ledCallbackType = iota
	ledCallbackHyperion
)

type lifxLEDCallback struct {
	Lights    []uint16
	RateLimit time.Duration
}

type lifxCallback struct {
	Power         bool
	PowerDuration time.Duration
	Color         lifxcommon.Color
	ColorDuration time.Duration
	Boblight      lifxLEDCallback
	Hyperion      lifxLEDCallback
	Lights        []string
	Groups        []string
}

type ledSync struct {
	lights map[uint64]closer
	sync.RWMutex
}

func (b *ledSync) add(l lifxcommon.Light, ledIDs []uint16, rateLimit time.Duration, kind ledCallbackType) {
	log.WithFields(log.Fields{
		`light`: l.ID(),
		`kind`:  kind,
	}).Debug(`Adding LED sync`)
	b.RLock()
	_, ok := b.lights[l.ID()]
	b.RUnlock()
	if ok {
		return
	}

	c := make(closer)
	b.Lock()
	b.lights[l.ID()] = c
	go b.sync(l, ledIDs, rateLimit, c, kind)
	b.Unlock()
}

func (b *ledSync) remove(l lifxcommon.Light) {
	b.Lock()
	log.WithField(`light`, l.ID()).Debug(`Removing LED sync`)
	if c, ok := b.lights[l.ID()]; ok {
		close(c)
		delete(b.lights, l.ID())
	}
	b.Unlock()
}

func (b *ledSync) stop() {
	b.Lock()
	for id, c := range b.lights {
		close(c)
		delete(b.lights, id)
	}
	b.Unlock()
}

func (b *ledSync) cancel(cb *lifxCallback) {
	log.WithField(`callback`, cb).Debug(`Cancelling LED sync`)
	if len(cb.Lights) == 0 && len(cb.Groups) == 0 {
		log.Debug(`Stopping LED sync`)
		b.stop()
	}
	if len(cb.Lights) > 0 {
		log.WithField(`lights`, cb.Lights).Debug(`Cancelling LED sync`)
		for _, label := range cb.Lights {
			light, err := client.GetLightByLabel(label)
			if err != nil {
				continue
			}
			leds.remove(light)
		}
	}
	if len(cb.Groups) > 0 {
		log.WithField(`groups`, cb.Groups).Debug(`Cancelling LED sync`)
		for _, label := range cb.Groups {
			group, err := client.GetGroupByLabel(label)
			if err != nil {
				continue
			}
			for _, light := range group.Lights() {
				leds.remove(light)
			}
		}
	}
}

func (b *ledSync) sync(l lifxcommon.Light, ledIDs []uint16, rateLimit time.Duration, c closer, kind ledCallbackType) {
	var (
		ledColor *common.Color
		err      error
	)
	ticker := time.NewTicker(time.Second / 20)
	for {
		select {
		case <-c:
			ticker.Stop()
			return
		case <-ticker.C:
			colors := make([]lifxcommon.Color, len(ledIDs))

			for i, id := range ledIDs {
				switch kind {
				case ledCallbackBoblight:
					ledColor, err = boblight.Lights.Get(id)
				case ledCallbackHyperion:
					ledColor, err = hyperion.Lights.Get(id)
				default:
					continue
				}
				if err != nil {
					continue
				}
				colors[i] = ledColor.ToLifx()
			}

			if err := l.SetColor(lifxcommon.AverageColor(colors...), rateLimit); err != nil {
				continue
			}
		}
	}
}

var (
	client *golifx.Client
	leds   = newLEDSync()
)

func initClient() error {
	var err error

	client, err = golifx.NewClient(&protocol.V2{Reliable: true})
	if err != nil {
		log.WithField(`error`, err).Error(`Creating LIFX client`)
		return err
	}

	return nil
}

// Connect establishes a LIFX client and performs device discovery
func Connect(cfg *config.Config) {
	if err := initClient(); err != nil {
		tick := time.Tick(2 * time.Second)
		done := make(closer)
		select {
		case <-done:
		case <-tick:
			err = initClient()
			if err == nil {
				close(done)
			}
		}
	}

	if err := client.SetDiscoveryInterval(30 * time.Second); err != nil {
		log.WithField(`error`, err).Fatal(`Failed setting lifx discovery interval`)
	}
	if cfg.LIFX.Timeout != nil && *cfg.LIFX.Timeout > 0 {
		client.SetTimeout(time.Duration(*cfg.LIFX.Timeout))
	}
	log.Info(`Initiated LIFX client`)
}

// Close closes the LIFX client
func Close() {
	_ = client.Close()
}

// Execute takes the callback and executes it via the LIFX protocol
func Execute(callback map[string]interface{}) {
	var (
		cb      lifxCallback
		decoder *mapstructure.Decoder
		err     error
	)

	log.WithField(`request`, callback).Debug(`Sending to LIFX`)

	decoder, err = mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: mapstructure.StringToTimeDurationHookFunc(),
		Result:     &cb,
	})
	if err != nil {
		log.WithField(`error`, err).Error(`Initializing LIFX callback decoder`)
		return
	}

	if err = decoder.Decode(callback); err != nil {
		log.WithFields(log.Fields{
			`request`: callback,
			`error`:   err,
		}).Error(`Decoding LIFX callback`)
		return
	}

	leds.cancel(&cb)

	if _, ok := callback[`power`]; ok {
		if len(cb.Lights) == 0 && len(cb.Groups) == 0 {
			if err = client.SetPowerDuration(cb.Power, cb.PowerDuration); err != nil {
				log.WithField(`error`, err).Error(`Setting power`)
			}
		}
		if len(cb.Lights) > 0 {
			for _, label := range cb.Lights {
				light, e := client.GetLightByLabel(label)
				if e != nil {
					log.WithFields(log.Fields{
						`label`: label,
						`error`: e,
					}).Error(`Finding light`)
					continue
				}
				if e = light.SetPowerDuration(cb.Power, cb.PowerDuration); e != nil {
					log.WithFields(log.Fields{
						`label`: label,
						`error`: e,
					}).Error(`Setting power`)
					continue
				}
			}
		}
		if len(cb.Groups) > 0 {
			for _, label := range cb.Groups {
				group, e := client.GetGroupByLabel(label)
				if e != nil {
					log.WithFields(log.Fields{
						`label`: label,
						`error`: e,
					}).Error(`Finding group`)
					continue
				}
				if e = group.SetPowerDuration(cb.Power, cb.PowerDuration); e != nil {
					log.WithFields(log.Fields{
						`label`: label,
						`error`: e,
					}).Error(`Setting power`)
					continue
				}
			}
		}
	}

	if _, ok := callback[`color`]; ok {
		if len(cb.Lights) == 0 && len(cb.Groups) == 0 {
			if err = client.SetColor(cb.Color, cb.ColorDuration); err != nil {
				log.WithField(`error`, err).Error(`Setting color`)
			}
		}
		if len(cb.Lights) > 0 {
			for _, label := range cb.Lights {
				light, err := client.GetLightByLabel(label)
				if err != nil {
					log.WithFields(log.Fields{
						`label`: label,
						`error`: err,
					}).Error(`Finding light`)
					continue
				}
				if err = light.SetColor(cb.Color, cb.ColorDuration); err != nil {
					log.WithFields(log.Fields{
						`label`: label,
						`error`: err,
					}).Error(`Setting color`)
					continue
				}
			}
		}
		if len(cb.Groups) > 0 {
			for _, label := range cb.Groups {
				group, err := client.GetGroupByLabel(label)
				if err != nil {
					log.WithFields(log.Fields{
						`label`: label,
						`error`: err,
					}).Error(`Finding group`)
					continue
				}
				if err = group.SetColor(cb.Color, cb.ColorDuration); err != nil {
					log.WithFields(log.Fields{
						`label`: label,
						`error`: err,
					}).Error(`Setting color`)
					continue
				}
			}
		}
	}

	if _, ok := callback[`boblight`]; ok {
		ledCallback(cb, ledCallbackBoblight)
	}

	if _, ok := callback[`hyperion`]; ok {
		ledCallback(cb, ledCallbackHyperion)
	}

}

func ledCallback(cb lifxCallback, kind ledCallbackType) {
	// Recommended LIFX device rate limit
	rateLimit := time.Second / 20
	// If requested rate limit is longer than minimum, use it
	var sourceLights []uint16
	switch kind {
	case ledCallbackBoblight:
		sourceLights = cb.Boblight.Lights
		if cb.Boblight.RateLimit > rateLimit {
			rateLimit = cb.Boblight.RateLimit
		}
	case ledCallbackHyperion:
		sourceLights = cb.Hyperion.Lights
		if cb.Hyperion.RateLimit > rateLimit {
			rateLimit = cb.Hyperion.RateLimit
		}
	default:
		return
	}

	// All lights
	if len(cb.Lights) == 0 && len(cb.Groups) == 0 {
		lights, err := client.GetLights()
		if err != nil {
			return
		}
		for _, light := range lights {
			leds.add(light, sourceLights, rateLimit, kind)
		}
	}

	// Explicit lights
	if len(cb.Lights) > 0 {
		for _, label := range cb.Lights {
			light, err := client.GetLightByLabel(label)
			if err != nil {
				log.WithFields(log.Fields{
					`label`: label,
					`error`: err,
				}).Error(`Finding light`)
				continue
			}
			leds.add(light, sourceLights, rateLimit, kind)
		}
	}

	// Groups
	if len(cb.Groups) > 0 {
		for _, label := range cb.Groups {
			group, err := client.GetGroupByLabel(label)
			if err != nil {
				log.WithFields(log.Fields{
					`label`: label,
					`error`: err,
				}).Error(`Finding group`)
				continue
			}
			for _, light := range group.Lights() {
				leds.add(light, sourceLights, rateLimit, kind)
			}
		}
	}
}

func newLEDSync() *ledSync {
	return &ledSync{
		lights: make(map[uint64]closer),
	}
}
