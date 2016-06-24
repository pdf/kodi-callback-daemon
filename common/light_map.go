package common

import (
	"fmt"
	"sync"
)

// LightMap provides storage for the current light state
type LightMap struct {
	lights map[uint16]*Color
	sync.RWMutex
}

// Get returns the last known state for the requested light ID, or error if
// unknown
func (l *LightMap) Get(id uint16) (color *Color, err error) {
	l.RLock()
	color, ok := l.lights[id]
	l.RUnlock()
	if !ok {
		return color, fmt.Errorf("Unkown light ID: %d", id)
	}
	return color, nil
}

// Set sets a light ID to the provided Color
func (l *LightMap) Set(id uint16, color *Color) {
	l.Lock()
	l.lights[id] = color
	l.Unlock()
}

// NewLightMap initializes a new LightMap
func NewLightMap() *LightMap {
	return &LightMap{
		lights: make(map[uint16]*Color),
	}
}
