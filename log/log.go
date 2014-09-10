package log

import (
	"errors"
	"fmt"

	"github.com/op/go-logging"
)

const (
	LOG_FORMAT = `[%{color}%{level}%{color:reset}] %{message}`
	LOG_MODULE = `xbmc-callback-daemon`
)

var (
	Logger *logging.Logger
)

func init() {
	logging.SetFormatter(logging.MustStringFormatter(LOG_FORMAT))
	Logger = logging.MustGetLogger(LOG_MODULE)
	logging.SetLevel(logging.INFO, LOG_MODULE)
}

// SetLogLevel adjusts the level of logger output
func SetLogLevel(level string) error {
	switch level {
	case `debug`, `DEBUG`:
		logging.SetLevel(logging.DEBUG, LOG_MODULE)
	case `notice`, `NOTICE`:
		logging.SetLevel(logging.NOTICE, LOG_MODULE)
	case `info`, `INFO`:
		logging.SetLevel(logging.INFO, LOG_MODULE)
	case `warning`, `WARNING`, `warn`, `WARN`:
		logging.SetLevel(logging.WARNING, LOG_MODULE)
	case `error`, `ERROR`:
		logging.SetLevel(logging.ERROR, LOG_MODULE)
	case `critical`, `CRITICAL`, `crit`, `CRIT`:
		logging.SetLevel(logging.CRITICAL, LOG_MODULE)
	default:
		return errors.New(fmt.Sprintf(`Unknown log level: %s`, level))
	}

	return nil
}
