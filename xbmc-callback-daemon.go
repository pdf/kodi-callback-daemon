package main

import (
	`github.com/pdf/xbmc-callback-daemon/hyperion`
	`github.com/pdf/xbmc-callback-daemon/xbmc`
)

func main() {
	xbmc.Connect(`127.0.0.1:9090`)
	defer xbmc.Close()
	hyperion.Connect(`127.0.0.1:19444`)
	defer hyperion.Close()
	notification := &xbmc.XBMCNotification{}

	hyperion.Send(hyperion.NewTransform())
	hyperion.Send(hyperion.NewEffect(`Rainbow swirl`))

	for {
		xbmc.Read(notification)

		switch notification.Method {
		case `System.OnQuit`, `System.OnRestart`:
			hyperion.Send(hyperion.NewColor([3]uint8{0, 0, 0}))

		case `Player.OnPlay`:
			t := hyperion.NewTransform()
			if notification.Params.Data.Item.Type == `song` {
				t.Transform.Gamma = [3]hyperion.Float64{0.8, 0.8, 0.8}
				t.Transform.SaturationGain = 2.0
				t.Transform.ValueGain = 2.0
			}
			hyperion.Send(t)
			hyperion.Send(hyperion.NewClear())

		case `Player.OnPause`:
			if notification.Params.Data.Item.Type == `song` {
				hyperion.Send(hyperion.NewTransform())
			}
			hyperion.Send(hyperion.NewEffect(`Red value mood blobs`))

		case `Player.OnStop`:
			hyperion.Send(hyperion.NewTransform())
			hyperion.Send(hyperion.NewEffect(`Rainbow swirl`))

		case `GUI.OnScreensaverActivated`:
			hyperion.Send(hyperion.NewColor([3]uint8{0, 0, 0}))

		case `GUI.OnScreensaverDeactivated`:
			hyperion.Send(hyperion.NewEffect(`Rainbow swirl`))

		}
	}
}
