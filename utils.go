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
			b.String(), exitStatus, strings.TrimSpace(out.String()))
	}

	return nil
}

// returns:
// 0 if equal
// 1 if v1 is grater than v2
// -1 if v1 is less than v2
func verCmp(v1 []int, v2 []int) int {
	maxLen := len(v1)
	if len(v2) > maxLen {
		maxLen = len(v2)
	}

	for i := 0; i < maxLen; i++ {
		switch {
		case i < len(v1) && i < len(v2) && v1[i] == v2[i]:
			continue
		case i < len(v1) && i < len(v2) && v1[i] > v2[i]:
			return 1
		case i < len(v1) && i < len(v2) && v1[i] < v2[i]:
			return -1
		case i < len(v1) && v1[i] > 0:
			return 1
		case i < len(v2) && v2[i] > 0:
			return -1
		}
	}

	return 0
}
