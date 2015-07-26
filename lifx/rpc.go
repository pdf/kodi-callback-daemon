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
}

var (
	client *golifx.Client
)

// Connect establishes a LIFX client and performs device discovery
func Connect(timeout *config.Timeout) {
	var err error

	golifx.SetLogger(log.New())
	client, err = golifx.NewClient(&protocol.V2{})
	if timeout != nil && *timeout > 0 {
		client.SetTimeout(time.Duration(*timeout))
	}
	for err != nil {
		log.WithField(`error`, err).Error(`Creating LIFX client`)
		log.Info(`Attempting rediscovery...`)
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
		if len(cb.Lights) == 0 {
			if err = client.SetPowerDuration(cb.Power, cb.PowerDuration); err != nil {
				log.WithField(`error`, err).Error(`Setting power`)
			}
		} else {
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
	}

	if _, ok := callback[`color`]; ok {
		if len(cb.Lights) == 0 {
			if err = client.SetColor(cb.Color, cb.ColorDuration); err != nil {
				log.WithField(`error`, err).Error(`Setting color`)
			}
		} else {
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
	}

}
