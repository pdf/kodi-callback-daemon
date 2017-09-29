package boblight

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"regexp"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/pdf/kodi-callback-daemon/common"
	"github.com/pdf/kodi-callback-daemon/config"
)

var (
	// Lights provides access to the current light state
	Lights = common.NewLightMap()

	lightRegexp = regexp.MustCompile(`^set light ([0-9]+) rgb ([0-9.]+) ([0-9.]+) ([0-9.]+)$`)
	quitChan    = make(chan struct{})
	cfg         *config.Config
)

type proxy struct {
	inputConn, outputConn *net.TCPConn
	inputAddr, outputAddr *net.TCPAddr
	doneChan              chan struct{}
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
			go newProxy(conn, inputAddr, outputAddr).start()
		}
	}
}

func (p *proxy) start() {
	var err error

	defer func() {
		if err = p.inputConn.Close(); err != nil {
			log.WithField(`error`, err).Error(`Failed closing boblight listener`)
		}
	}()

	p.outputConn, err = net.DialTCP(`tcp`, nil, p.outputAddr)
	if err != nil {
		log.WithField(`error`, err).Error(`Failed connecting to boblight output address`)
		return
	}

	go p.proxy(true)
	go p.proxy(false)
	select {
	case <-p.doneChan:
		return
	case <-quitChan:
		return
	}
}

func (p *proxy) proxy(input bool) {
	for {
		select {
		case <-p.doneChan:
			return
		case <-quitChan:
			return
		default:
			var src, dst *net.TCPConn
			if input {
				src, dst = p.inputConn, p.outputConn
			} else {
				src, dst = p.outputConn, p.inputConn
			}

			data := make([]byte, 0xffff)
			n, err := src.Read(data)
			if err != nil {
				if err != io.EOF {
					log.WithField(`error`, err).Warn(`Boblight proxy read failed`)
				}
				p.done()
				return
			}

			if _, err = dst.Write(data[:n]); err != nil {
				if err != io.EOF {
					log.WithField(`error`, err).Warn(`Boblight proxy write failed`)
				}
				p.done()
				return
			}

			if input {
				parse(data[:n])
			}
		}
	}
}

func (p *proxy) done() {
	select {
	case <-p.doneChan:
		return
	default:
		close(p.doneChan)
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
		Lights.Set(
			uint16(id),
			&common.Color{colorful.Color{
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

func newProxy(inputConn *net.TCPConn, inputAddr, outputAddr *net.TCPAddr) *proxy {
	return &proxy{
		inputConn:  inputConn,
		inputAddr:  inputAddr,
		outputAddr: outputAddr,
		doneChan:   make(chan struct{}),
	}
}
