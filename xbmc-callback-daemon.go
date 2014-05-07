package main

import (
	`encoding/json`
	`log`
	`net`
	`os/exec`
)

type Item struct {
	Type string
}

type Message struct {
	Method string
	Params struct {
		Data struct {
			Item *Item
		}
	}
}

func main() {
	conn, err := net.Dial(`tcp`, `127.0.0.1:9090`)
	if err != nil {
		log.Println(`Couldn't connect to XBMC`)
		return
	}
	dec := json.NewDecoder(conn)

	for {
		m := &Message{}

		if err := dec.Decode(&m); err != nil {
			log.Println(err)
			return
		}

		switch m.Method {
		case `Player.OnPlay`:
			onPlay(m.Params.Data.Item.Type)
		case `Player.OnPause`:
			onPause()
		case `Player.OnStop`:
			onStop()
		case `GUI.OnScreensaverActivated`:
			onScreensaverActivated()
		case `GUI.OnScreensaverDeactivated`:
			onScreensaverDeactivated()
		}
	}
}

func onPlay(t string) {
	switch t {
	case `music`:
		if _, err := exec.Command(`/usr/local/bin/hyperion-remote`, `--gamma`, `0.8 0.8 0.8`, `--value`, `2.0`, `--saturation`, `2.0`).Output(); err != nil {
			log.Println(err)
		}
	default:
		defaultGamma()
	}

	if _, err := exec.Command(`/usr/local/bin/hyperion-remote`, `--priority`, `86`, `--clear`).Output(); err != nil {
		log.Println(err)
	}
}

func onPause() {
	defaultGamma()
	if _, err := exec.Command(`/usr/local/bin/hyperion-remote`, `--priority`, `86`, `--effect`, `Red value mood blobs`).Output(); err != nil {
		log.Println(err)
	}
}

func onStop() {
	defaultGamma()
	if _, err := exec.Command(`/usr/local/bin/hyperion-remote`, `--priority`, `86`, `--effect`, `Rainbow swirl`).Output(); err != nil {
		log.Println(err)
	}
}

func onScreensaverActivated() {
	if _, err := exec.Command(`/usr/local/bin/hyperion-remote`, `--priority`, `86`, `--color`, `000000`).Output(); err != nil {
		log.Println(err)
	}
}

func onScreensaverDeactivated() {
	defaultGamma()
	if _, err := exec.Command(`/usr/local/bin/hyperion-remote`, `--priority`, `86`, `--effect`, `Rainbow swirl`).Output(); err != nil {
		log.Println(err)
	}
}

func defaultGamma() {
	if _, err := exec.Command(`/usr/local/bin/hyperion-remote`, `--gamma`, `2.2 2.2 2.8`, `--value`, `1.0`, `--saturation`, `1.0`).Output(); err != nil {
		log.Println(err)
	}
}
