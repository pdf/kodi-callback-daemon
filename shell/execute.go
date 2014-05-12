package shell

import (
	`fmt`
	`os/exec`

	`github.com/pdf/xbmc-callback-daemon/logger`
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

	logger.Debug(`Executing shell command: `, cmd)
	if callback[`background`] != nil && callback[`background`] == true {
		if err = cmd.Start(); err != nil {
			logger.Warn(fmt.Sprintf("Failure executing command (%v): %v", cmd, err))
		}
	} else {
		if err = cmd.Run(); err != nil {
			logger.Warn(fmt.Sprintf("Failure executing command (%v): %v", cmd, err))
		}
	}
}
