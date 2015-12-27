package main

import (
	"errors"
	"fmt"
	"os/exec"
)

func checkDeps() error {
	var missingDeps []string

	if _, err := exec.LookPath("hostapd"); err != nil {
		missingDeps = append(missingDeps, "hostapd")
	}

	if _, err := exec.LookPath("iptables"); err != nil {
		missingDeps = append(missingDeps, "iptables")
	}

	if _, err := exec.LookPath("dnsmasq"); err != nil {
		missingDeps = append(missingDeps, "dnsmasq")
	}

	if _, err := exec.LookPath("iw"); err != nil {
		missingDeps = append(missingDeps, "iw")
	}

	if _, err := exec.LookPath("ip"); err != nil {
		missingDeps = append(missingDeps, "iproute2")
	}

	if len(missingDeps) == 0 {
		return nil
	}

	errStr := "The following dependencies are not found:"
	for _, dep := range missingDeps {
		errStr += fmt.Sprintf("\n  %s", dep)
	}

	return errors.New(errStr)
}
