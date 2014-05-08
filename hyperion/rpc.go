package hyperion

import (
	`encoding/json`
	`log`
	`net`
	`reflect`
	`strconv`
)

const (
	DEFAULT_PRIORITY = 86
)

var conn net.Conn
var Encoder *json.Encoder

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

// Hyperion `clear` command
type Clear struct {
	Command  string `json:"command"`
	Priority uint8  `json:"priority,omitempty"`
}

// `clear` defaults
func NewClear() *Clear {
	return &Clear{Command: `clear`, Priority: DEFAULT_PRIORITY}
}

// Hyperion `transform` command
type Transform struct {
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
func NewTransform() *Transform {
	h := &Transform{}
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
type Effect struct {
	Command  string `json:"command"`
	Priority uint8  `json:"priority,omitempty"`
	Duration uint16 `json:"duration,omitempty"`
	Effect   struct {
		Name string `json:"name"`
		Args string `json:"args,omitempty"`
	} `json:"effect"`
}

func NewEffect(name string) *Effect {
	h := &Effect{Command: `effect`, Priority: DEFAULT_PRIORITY}
	h.Effect.Name = name
	return h
}

// Hyperion `color` command
type Color struct {
	Command  string   `json:"command"`
	Priority uint8    `json:"priority,omitempty"`
	Duration uint16   `json:"duration,omitempty"`
	Color    [3]uint8 `json:"color"`
}

// `color` defaults
func NewColor(color [3]uint8) *Color {
	h := &Color{Command: `color`, Priority: DEFAULT_PRIORITY}
	h.Color = color
	return h
}

// Initialize connection
func Connect(address string) {
	conn, err := net.Dial(`tcp`, address)
	if err != nil {
		log.Panicln(`[ERROR] Connecting to Hyperion: `, err)
	} else {
		log.Println(`[INFO] Connected to Hyperion`)
	}
	Encoder = json.NewEncoder(conn)
}

// Close Hyperion connection
func Close() {
	conn.Close()
}

// Send encoded data to Hyperion
func Send(command interface{}) {
	if err := Encoder.Encode(&command); err != nil {
		log.Println(`[ERROR] Sending to Hyperion: `, err)
	}
}
