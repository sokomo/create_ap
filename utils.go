package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
)

func runCmd(name string, args ...string) error {
	var out bytes.Buffer

	cmd := exec.Command(name, args...)
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	if err != nil {
		var b bytes.Buffer
		var exitStatus int

		for _, arg := range cmd.Args {
			if b.Len() > 0 {
				b.WriteByte(' ')
			}

			f := strings.Contains(arg, " ")
			if f {
				b.WriteByte('\'')
			}

			b.WriteString(arg)

			if f {
				b.WriteByte('\'')
			}
		}

		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				exitStatus = status.ExitStatus()
			}
		}

		return fmt.Errorf("cmd: %s\nexit status: %d\nout: %s",
			b.String(), exitStatus, strings.Trim(out.String(), " \t\n"))
	}

	return nil
}
