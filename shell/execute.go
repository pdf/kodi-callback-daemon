package shell

import (
	`os/exec`

	. `github.com/pdf/xbmc-callback-daemon/log`
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

	Logger.Debug(`Executing shell command: %v`, cmd)
	if callback[`background`] != nil && callback[`background`] == true {
		if err = cmd.Start(); err != nil {
			Logger.Warning(`Failure executing command (%v): %v`, cmd, err)
		}
		go func() {
			if err = cmd.Wait(); err != nil {
				Logger.Warning(`Background command exited with error (%v): %v`, cmd, err)
			}
		}()
	} else {
		if err = cmd.Run(); err != nil {
			Logger.Warning(`Failure executing command (%v): %v`, cmd, err)
		}
	}
}
