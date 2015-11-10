package boblight

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"regexp"
	"strconv"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/pdf/golifx/common"
	"github.com/pdf/kodi-callback-daemon/config"
)

var (
	// Lights provides access to the current light state
	Lights = newLightMap()

	lightRegexp = regexp.MustCompile(`^set light ([0-9]+) rgb ([0-9.]+) ([0-9.]+) ([0-9.]+)$`)
	quitChan    = make(chan struct{})
	cfg         *config.Config
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

func (l *LightMap) set(id uint16, color *Color) {
	l.Lock()
	l.lights[id] = color
	l.Unlock()
}

func listen(inputAddr, outputAddr *net.TCPAddr) {
	listener, err := net.ListenTCP(`tcp`, inputAddr)
	if err != nil {
		log.WithField(`error`, err).Fatal(`Could not open requested boblight input address`)
	}

	for {
		select {
		case <-quitChan:
			if err := listener.Close(); err != nil {
				panic(err)
			}
			return
		default:
			conn, err := listener.AcceptTCP()
			if err != nil {
				log.WithField(`error`, err).Error(`Failed accepting boblight connection`)
				continue
			}
			go start(conn, inputAddr, outputAddr)
		}
	}
}

func start(inputConn *net.TCPConn, inputAddr, outputAddr *net.TCPAddr) {
	defer func() {
		if err := inputConn.Close(); err != nil {
			log.WithField(`error`, err).Error(`Failed closing boblight listener`)
		}
	}()

	outputConn, err := net.DialTCP(`tcp`, nil, outputAddr)
	if err != nil {
		log.WithField(`error`, err).Error(`Failed connecting to boblight output address`)
		return
	}

	done := make(chan struct{})
	go proxy(inputConn, outputConn, true, done)
	go proxy(outputConn, inputConn, false, done)
	select {
	case <-done:
		return
	case <-quitChan:
		return
	}
}

func proxy(input, output *net.TCPConn, shouldParse bool, done chan struct{}) {
	for {
		select {
		case <-done:
			return
		case <-quitChan:
			return
		default:
			data := make([]byte, 0xffff)
			n, err := input.Read(data)
			if err != nil {
				if err != io.EOF {
					log.WithField(`error`, err).Warn(`Boblight proxy read failed`)
				}
				close(done)
				return
			}

			if _, err = output.Write(data[:n]); err != nil {
				if err != io.EOF {
					log.WithField(`error`, err).Warn(`Boblight proxy write failed`)
				}
				close(done)
				return
			}

			if shouldParse {
				parse(data[:n])
			}
		}
	}
}

func parse(data []byte) {
	buf := bytes.NewBuffer(data)

	scanner := bufio.NewScanner(buf)
	for scanner.Scan() {
		line := scanner.Text()
		match := lightRegexp.FindStringSubmatch(line)
		if match == nil {
			continue
		}
		var (
			id      uint64
			r, g, b float64
			err     error
		)
		id, err = strconv.ParseUint(match[1], 10, 16)
		if err != nil {
			continue
		}
		r, err = strconv.ParseFloat(match[2], 64)
		if err != nil {
			continue
		}
		g, err = strconv.ParseFloat(match[3], 64)
		if err != nil {
			continue
		}
		b, err = strconv.ParseFloat(match[4], 64)
		if err != nil {
			continue
		}
		Lights.set(
			uint16(id),
			&Color{colorful.Color{
				R: r,
				G: g,
				B: b,
			}},
		)
	}
}

// Connect initializes the Boblight proxy
func Connect(conf *config.Config) {
	cfg = conf

	if cfg.Boblight.Input == nil {
		log.Fatal(`No input configuration for boblight backend`)
	}
	inputAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", cfg.Boblight.Input.Address, cfg.Boblight.Input.Port))
	if err != nil {
		log.WithField(`error`, err).Fatal(`Could not parse requested boblight input address`)
	}
	outputAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", cfg.Boblight.Output.Address, cfg.Boblight.Output.Port))
	if err != nil {
		log.WithField(`error`, err).Fatal(`Could not parse requested boblight output address`)
	}

	go listen(inputAddr, outputAddr)
	log.Info(`Initiated Boblight proxy`)
}

// Close Boblight proxy
func Close() {
	close(quitChan)
}

// newLightMap initializes a new LightMap
func newLightMap() *LightMap {
	return &LightMap{
		lights: make(map[uint16]*Color),
	}
}
