package main

import (
	`encoding/json`
	`io`
	`log`
	`net`
	`reflect`
	`strconv`
)

const (
	DEFAULT_PRIORITY = 86
)

// Custom Float64 class to ensure marshalling as floats for round numbers
type Float64 float64

func (f Float64) MarshalText() ([]byte, error) {
	v := reflect.ValueOf(f)
	return []byte(strconv.FormatFloat(v.Float(), 'f', 6, 64)), nil
}

func (f Float64) MarshalJSON() ([]byte, error) {
	v := reflect.ValueOf(f)
	return []byte(strconv.FormatFloat(v.Float(), 'f', 6, 64)), nil
}

// XBMC server->client notification
type XBMCNotification struct {
	Method string
	Params struct {
		Data struct {
			Item *struct {
				Type string
			}
		}
	}
}

// Hyperion `clear` command
type HyperionClear struct {
	Command  string `json:"command"`
	Priority uint8  `json:"priority,omitempty"`
}

// `clear` defaults
func NewHyperionClear() *HyperionClear {
	return &HyperionClear{Command: `clear`, Priority: DEFAULT_PRIORITY}
}

// Hyperion `transform` command
type HyperionTransform struct {
	Command   string `json:"command"`
	Priority  uint8  `json:"priority,omitempty"`
	Transform struct {
		Id             string     `json:"id,omitempty"`
		SaturationGain Float64    `json:"saturationGain,omitempty"`
		ValueGain      Float64    `json:"valueGain,omitempty"`
		Threshold      [3]Float64 `json:"threshold,omitempty"`
		Gamma          [3]Float64 `json:"gamma,omitempty"`
		Blacklevel     [3]Float64 `json:"blacklevel,omitempty"`
		Whitelevel     [3]Float64 `json:"whitelevel,omitempty"`
	} `json:"transform"`
}

// `transform` defaults
func NewHyperionTransform() *HyperionTransform {
	h := &HyperionTransform{}
	h.Command = `transform`
	h.Transform.SaturationGain = 1.0
	h.Transform.ValueGain = 1.0
	h.Transform.Threshold = [3]Float64{0.1, 0.1, 0.1}
	h.Transform.Gamma = [3]Float64{2.2, 2.2, 2.8}
	h.Transform.Blacklevel = [3]Float64{0.0, 0.0, 0.0}
	h.Transform.Whitelevel = [3]Float64{1.0, 1.0, 1.0}
	return h
}

// Hyperion `effect` command
type HyperionEffect struct {
	Command  string `json:"command"`
	Priority uint8  `json:"priority,omitempty"`
	Duration uint16 `json:"duration,omitempty"`
	Effect   struct {
		Name string `json:"name"`
		Args string `json:"args,omitempty"`
	} `json:"effect"`
}

// `effect` defaults
func NewHyperionEffect(name string) *HyperionEffect {
	h := &HyperionEffect{Command: `effect`, Priority: DEFAULT_PRIORITY}
	h.Effect.Name = name
	return h
}

// Hyperion `color` command
type HyperionColor struct {
	Command  string   `json:"command"`
	Priority uint8    `json:"priority,omitempty"`
	Duration uint16   `json:"duration,omitempty"`
	Color    [3]uint8 `json:"color"`
}

// `color` defaults
func NewHyperionColor(color [3]uint8) *HyperionColor {
	h := &HyperionColor{Command: `color`, Priority: DEFAULT_PRIORITY}
	h.Color = color
	return h
}

func main() {
	xbmc, err := net.Dial(`tcp`, `127.0.0.1:9090`)
	if err != nil {
		log.Println(`Couldn't connect to XBMC`)
		return
	}
	defer xbmc.Close()
	dec := json.NewDecoder(xbmc)

	hyperion, err := net.Dial(`tcp`, `127.0.0.1:19444`)
	if err != nil {
		log.Println(`Couldn't connect to Hyperion`)
		return
	}
	defer hyperion.Close()
	enc := json.NewEncoder(hyperion)

	// Set rainbow swirl on startup
	t := NewHyperionTransform()
	if err := enc.Encode(&t); err != nil {
		log.Println(err)
		return
	}
	e := NewHyperionEffect(`Rainbow swirl`)
	if err := enc.Encode(&e); err != nil {
		log.Println(err)
		return
	}

	for {
		m := &XBMCNotification{}

		if err := dec.Decode(&m); err == io.EOF {
			log.Println(err)
			return
		}

		switch m.Method {
		case `System.OnQuit`, `System.OnRestart`:
			c := NewHyperionColor([3]uint8{0, 0, 0})
			if err := enc.Encode(&c); err != nil {
				log.Println(err)
				return
			}

		case `Player.OnPlay`:
			t := NewHyperionTransform()
			if m.Params.Data.Item.Type == `song` {
				t.Transform.Gamma = [3]Float64{0.8, 0.8, 0.8}
				t.Transform.SaturationGain = 2.0
				t.Transform.ValueGain = 2.0
			}
			if err := enc.Encode(&t); err != nil {
				log.Println(err)
				return
			}
			c := NewHyperionClear()
			if err := enc.Encode(&c); err != nil {
				log.Println(err)
				return
			}

		case `Player.OnPause`:
			if m.Params.Data.Item.Type == `song` {
				t := NewHyperionTransform()
				if err := enc.Encode(&t); err != nil {
					log.Println(err)
					return
				}
			}
			e := NewHyperionEffect(`Red value mood blobs`)
			if err := enc.Encode(&e); err != nil {
				log.Println(err)
				return
			}

		case `Player.OnStop`:
			t := NewHyperionTransform()
			if err := enc.Encode(&t); err != nil {
				log.Println(err)
				return
			}
			e := NewHyperionEffect(`Rainbow swirl`)
			if err := enc.Encode(&e); err != nil {
				log.Println(err)
				return
			}

		case `GUI.OnScreensaverActivated`:
			c := NewHyperionColor([3]uint8{0, 0, 0})
			if err := enc.Encode(&c); err != nil {
				log.Println(err)
				return
			}

		case `GUI.OnScreensaverDeactivated`:
			e := NewHyperionEffect(`Rainbow swirl`)
			if err := enc.Encode(&e); err != nil {
				log.Println(err)
				return
			}
		}
	}
}
