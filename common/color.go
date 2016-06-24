package common

import (
	"github.com/lucasb-eyer/go-colorful"
	"github.com/pdf/golifx/common"
)

// Color in boblight format (components in range 0.0 - 1.0)
type Color struct {
	colorful.Color
}

// ToLifx returns a golifx-compatible color
func (c *Color) ToLifx() (lifxColor common.Color) {
	h, s, v := c.Hsv()
	lifxColor.Hue = uint16((h / 360) * 65535)
	lifxColor.Saturation = uint16(s * 65535)
	lifxColor.Brightness = uint16(v * 65535)
	lifxColor.Kelvin = 0xffff
	return lifxColor
}
