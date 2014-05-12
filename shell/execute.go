package shell

import (
	`log`
	`os/exec`
)

// Execute takes the callback command and argument, and spawns the process.
func Execute(callback map[string]interface{}) {
	var (
		err error
	)

	bin := callback[`command`].(string)
	args := make([]string, 0)

	if callback[`arguments`] != nil {
		cbArguments := callback[`arguments`].([]interface{})
		for i := range cbArguments {
			args = append(args, cbArguments[i].(string))
		}
	}

	cmd := exec.Command(bin, args...)

	if callback[`background`] != nil && callback[`background`] == true {
		if err = cmd.Start(); err != nil {
			log.Printf("[WARNING] Failure executing command (%v): %v\n", cmd, err)
		}
	} else {
		if err = cmd.Run(); err != nil {
			log.Printf("[WARNING] Failure executing command (%v): %v\n%v\n", cmd, err)
		}
	}
}
