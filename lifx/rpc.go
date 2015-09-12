package lifx

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/mitchellh/mapstructure"

	"github.com/pdf/golifx"
	"github.com/pdf/golifx/common"
	"github.com/pdf/golifx/protocol"
	"github.com/pdf/kodi-callback-daemon/config"
)

type lifxCallback struct {
	Power         bool
	PowerDuration time.Duration
	Color         common.Color
	ColorDuration time.Duration
	Lights        []string
	Groups        []string
}

var (
	client *golifx.Client
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
func Connect(cfg config.Config) {
	logger := log.New()

	if cfg.Debug != nil && *cfg.Debug {
		logger.Level = log.DebugLevel
	}
	golifx.SetLogger(logger)

	if err := initClient(); err != nil {
		tick := time.Tick(2 * time.Second)
		done := make(chan bool)
		select {
		case <-done:
		case <-tick:
			err = initClient()
			if err == nil {
				done <- true
			}
		}
	}

	client.SetDiscoveryInterval(30 * time.Second)
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

	if _, ok := callback[`power`]; ok {
		if len(cb.Lights) == 0 && len(cb.Groups) == 0 {
			if err = client.SetPowerDuration(cb.Power, cb.PowerDuration); err != nil {
				log.WithField(`error`, err).Error(`Setting power`)
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
				if err = light.SetPowerDuration(cb.Power, cb.PowerDuration); err != nil {
					log.WithFields(log.Fields{
						`label`: label,
						`error`: err,
					}).Error(`Setting power`)
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
				if err = group.SetPowerDuration(cb.Power, cb.PowerDuration); err != nil {
					log.WithFields(log.Fields{
						`label`: label,
						`error`: err,
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

}
