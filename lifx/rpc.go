package lifx

import (
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/mitchellh/mapstructure"

	"github.com/pdf/golifx"
	"github.com/pdf/golifx/common"
	"github.com/pdf/golifx/protocol"
	"github.com/pdf/kodi-callback-daemon/boblight"
	"github.com/pdf/kodi-callback-daemon/config"
)

type closer chan struct{}

type lifxBoblightCallback struct {
	Lights    []uint16
	RateLimit time.Duration
}

type lifxCallback struct {
	Power         bool
	PowerDuration time.Duration
	Color         common.Color
	ColorDuration time.Duration
	Boblight      lifxBoblightCallback
	Lights        []string
	Groups        []string
}

type boblightSync struct {
	lights map[uint64]closer
	sync.RWMutex
}

func (b *boblightSync) add(l common.Light, boblightIDs []uint16, rateLimit time.Duration) {
	log.WithField(`light`, l.ID()).Debug(`Adding boblight sync`)
	b.RLock()
	_, ok := b.lights[l.ID()]
	b.RUnlock()
	if ok {
		return
	}

	c := make(closer)
	b.Lock()
	b.lights[l.ID()] = c
	go b.sync(l, boblightIDs, rateLimit, c)
	b.Unlock()
}

func (b *boblightSync) remove(l common.Light) {
	b.Lock()
	log.WithField(`light`, l.ID()).Debug(`Removing boblight sync`)
	if c, ok := b.lights[l.ID()]; ok {
		close(c)
		delete(b.lights, l.ID())
	}
	b.Unlock()
}

func (b *boblightSync) stop() {
	b.Lock()
	for id, c := range b.lights {
		close(c)
		delete(b.lights, id)
	}
	b.Unlock()
}

func (b *boblightSync) cancel(cb *lifxCallback) {
	log.WithField(`callback`, cb).Debug(`Cancelling boblight sync`)
	if len(cb.Lights) == 0 && len(cb.Groups) == 0 {
		log.Debug(`Stopping boblight sync`)
		b.stop()
	}
	if len(cb.Lights) > 0 {
		log.WithField(`lights`, cb.Lights).Debug(`Cancelling boblight sync`)
		for _, label := range cb.Lights {
			light, err := client.GetLightByLabel(label)
			if err != nil {
				continue
			}
			boblights.remove(light)
		}
	}
	if len(cb.Groups) > 0 {
		log.WithField(`groups`, cb.Groups).Debug(`Cancelling boblight sync`)
		for _, label := range cb.Groups {
			group, err := client.GetGroupByLabel(label)
			if err != nil {
				continue
			}
			for _, light := range group.Lights() {
				boblights.remove(light)
			}
		}
	}
}

func (b *boblightSync) sync(l common.Light, boblightIDs []uint16, rateLimit time.Duration, c closer) {
	ticker := time.NewTicker(time.Second / 20)
	for {
		select {
		case <-c:
			ticker.Stop()
			return
		case <-ticker.C:
			var (
				hueSum, satSum, brightSum, kelvSum uint64
				color                              common.Color
			)

			for _, id := range boblightIDs {
				bColor, err := boblight.Lights.Get(id)
				if err != nil {
					continue
				}
				colr := bColor.ToLifx()

				hueSum += uint64(colr.Hue)
				satSum += uint64(colr.Saturation)
				brightSum += uint64(colr.Brightness)
				kelvSum += uint64(colr.Kelvin)
			}

			color.Hue = uint16(hueSum / uint64(len(boblightIDs)))
			color.Saturation = uint16(satSum / uint64(len(boblightIDs)))
			color.Brightness = uint16(brightSum / uint64(len(boblightIDs)))
			color.Kelvin = uint16(kelvSum / uint64(len(boblightIDs)))

			if err := l.SetColor(color, rateLimit); err != nil {
				continue
			}
		}
	}
}

var (
	client    *golifx.Client
	boblights = newBoblightSync()
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
	logger := log.New()

	if cfg.Debug != nil && *cfg.Debug {
		logger.Level = log.DebugLevel
	}
	golifx.SetLogger(logger)

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

	boblights.cancel(&cb)

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
		// Recommended LIFX device rate limit
		rateLimit := time.Second / 20
		// If requested rate limit is longer than minimum, use it
		if cb.Boblight.RateLimit > rateLimit {
			rateLimit = cb.Boblight.RateLimit
		}
		if len(cb.Lights) == 0 && len(cb.Groups) == 0 {
			lights, err := client.GetLights()
			if err != nil {
				return
			}
			for _, light := range lights {
				boblights.add(light, cb.Boblight.Lights, rateLimit)
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
				boblights.add(light, cb.Boblight.Lights, rateLimit)
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
				for _, light := range group.Lights() {
					boblights.add(light, cb.Boblight.Lights, rateLimit)
				}
			}
		}
	}

}

func newBoblightSync() *boblightSync {
	return &boblightSync{
		lights: make(map[uint64]closer),
	}
}
