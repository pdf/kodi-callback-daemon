package hyperion

import (
	"net"

	log "github.com/Sirupsen/logrus"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/pdf/kodi-callback-daemon/common"
)

var (
	// Lights provides access to the current light state
	Lights = common.NewLightMap()
)

func handleUDP(conn *net.UDPConn) error {
	log.Info(`Accepted Hyperion UDP connection`)
	defer func() {
		if err := conn.Close(); err != nil {
			log.WithField(`error`, err).Warn(`Failed closing Hyperion server connection`)
		}
	}()

	buf := make([]byte, 4096)

	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			return err
		}
		for i := 0; i <= n-3; i += 3 {
			Lights.Set(uint16(i/3), &common.Color{
				Color: colorful.Color{
					R: float64(buf[i]) / 255.0,
					G: float64(buf[i+1]) / 255.0,
					B: float64(buf[i+2]) / 255.0,
				},
			})
		}
	}
}

// listen Hyperion UDP LED data
func listen(address string) error {
	addr, err := net.ResolveUDPAddr(`udp`, address)
	if err != nil {
		return err
	}
	conn, err := net.ListenUDP(`udp`, addr)
	if err != nil {
		return err
	}
	defer func() {
		if err = conn.Close(); err != nil {
			log.WithField(`error`, err).Warn(`Failed closing Hyperion listener`)
		}
	}()

	return handleUDP(conn)
}
