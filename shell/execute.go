package shell

import (
	`os/exec`

	log "github.com/Sirupsen/logrus"
)

// Execute takes the callback command and argument, and spawns the process.
func Execute(callback map[string]interface{}) {
	var (
		err error
	)

	bin := callback[`command`].(string)
	args := make([]string, 0)

	// Build arguments slice
	if callback[`arguments`] != nil {
		cbArguments := callback[`arguments`].([]interface{})
		for i := range cbArguments {
			args = append(args, cbArguments[i].(string))
		}
	}

	cmd := exec.Command(bin, args...)

	log.WithField(`command`, cmd).Debug(`Executing shell command`)
	if callback[`background`] != nil && callback[`background`] == true {
		if err = cmd.Start(); err != nil {
			log.WithFields(log.Fields{
				`command`: cmd,
				`error`:   err,
			}).Warn(`Failure executing shell command`)
		}
		go func() {
			if err = cmd.Wait(); err != nil {
				log.WithFields(log.Fields{
					`command`: cmd,
					`error`:   err,
				}).Warn(`Background shell command exited with error`)
			}
		}()
	} else {
		if err = cmd.Run(); err != nil {
			log.WithFields(log.Fields{
				`command`: cmd,
				`error`:   err,
			}).Warn(`Failure executing command`)
		}
	}
}
