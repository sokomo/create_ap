package main

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	networkManagerConf = "/etc/NetworkManager/NetworkManager.conf"
)

func init() {
	// Make sure command outputs are in english, otherwise parsing will fail.
	os.Setenv("LC_ALL", "C")
}

func hasNetworkManager() bool {
	_, err := exec.LookPath("nmcli")
	return err == nil
}

func hasOldNetworkManager() bool {
	if !hasNetworkManager() {
		return false
	}

	v1, v2, v3 := networkManagerVersion()
	return verCmp([]int{v1, v2, v3}, []int{0, 9, 9}) == -1
}

func networkManagerVersion() (v1 int, v2 int, v3 int) {
	var out bytes.Buffer

	cmd := exec.Command("nmcli", "-v")
	cmd.Stdout = &out
	cmd.Stderr = &out
	cmd.Run()

	re := regexp.MustCompile("([0-9]+)\\.([0-9]+)(\\.([0-9]+))?")
	res := re.FindStringSubmatch(out.String())

	if len(res) == 5 {
		v1, _ = strconv.Atoi(res[1])
		v2, _ = strconv.Atoi(res[2])

		if res[4] != "" {
			v3, _ = strconv.Atoi(res[4])
		}
	}

	return
}

func networkManagerRunning() bool {
	var cmd *exec.Cmd
	var out bytes.Buffer

	if !hasNetworkManager() {
		return false
	}

	if hasOldNetworkManager() {
		cmd = exec.Command("nmcli", "-t", "-f", "RUNNING", "nm")
	} else {
		cmd = exec.Command("nmcli", "-t", "-f", "RUNNING", "g")
	}

	cmd.Stdout = &out
	cmd.Stderr = &out
	cmd.Run()

	return strings.TrimSpace(out.String()) == "running"
}

func networkManagerReadConf() (lines []string, keyfileIndex int, unmanagedIndex int, err error) {
	f, err := os.Open(networkManagerConf)
	if err != nil {
		return nil, -1, -1, err
	}
	defer f.Close()

	keyfileIndex = -1
	unmanagedIndex = -1
	bio := bufio.NewReader(f)

	for i := 0; ; i++ {
		line, err := bio.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, -1, -1, err
		}

		line = strings.TrimSpace(line)

		switch {
		case strings.HasPrefix(line, "[keyfile]"):
			keyfileIndex = i
		case strings.HasPrefix(line, "unmanaged-devices="):
			unmanagedIndex = i
		}

		lines = append(lines, line)
	}

	err = nil
	return
}

func networkManagerWriteConf(lines []string) error {
	f, err := os.OpenFile(networkManagerConf, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, s := range lines {
		f.WriteString(s + "\n")
	}

	return nil
}

func networkManagerAddUnmanaged(ifname string) error {
	var unmanaged []string
	var mac string

	lines, keyfileIndex, unmanagedIndex, err := networkManagerReadConf()
	if err != nil {
		return err
	}

	if unmanagedIndex != -1 {
		s := lines[unmanagedIndex]
		s = s[len("unmanaged-devices="):]
		s = strings.Replace(s, ",", ";", -1)
		unmanaged = strings.Split(s, ";")
	}

	if iface, err := net.InterfaceByName(ifname); err == nil {
		mac = iface.HardwareAddr.String()
	}

	for _, s := range unmanaged {
		if s == "interface-name:"+ifname || s == "mac:"+mac {
			return nil
		}
	}

	if hasOldNetworkManager() {
		unmanaged = append(unmanaged, "mac:"+mac)
	} else {
		unmanaged = append(unmanaged, "interface-name:"+ifname)
	}

	if keyfileIndex == -1 {
		if lines[len(lines)-1] != "" {
			lines = append(lines, "")
		}
		lines = append(lines, "[keyfile]")
		keyfileIndex = len(lines) - 1
	}

	s := "unmanaged-devices=" + strings.Join(unmanaged, ";")

	if unmanagedIndex != -1 {
		lines[unmanagedIndex] = s
	} else {
		lines = append(lines[:keyfileIndex+1], append([]string{s}, lines[keyfileIndex+1:]...)...)
	}

	return networkManagerWriteConf(lines)
}

func networkManagerRemoveUnmanaged(ifname string) error {
	var unmanaged []string
	var mac string

	lines, _, unmanagedIndex, err := networkManagerReadConf()
	if err != nil {
		return err
	}

	if unmanagedIndex == -1 {
		return nil
	}

	s := lines[unmanagedIndex]
	s = s[len("unmanaged-devices="):]
	s = strings.Replace(s, ",", ";", -1)
	unmanaged = strings.Split(s, ";")

	if iface, err := net.InterfaceByName(ifname); err == nil {
		mac = iface.HardwareAddr.String()
	}

	for i := 0; i < len(unmanaged); {
		if unmanaged[i] == "interface-name:"+ifname || unmanaged[i] == "mac:"+mac {
			unmanaged = append(unmanaged[:i], unmanaged[i+1:]...)
			continue
		}
		i++
	}

	if len(unmanaged) > 0 {
		lines[unmanagedIndex] = "unmanaged-devices=" + strings.Join(unmanaged, ";")
	} else {
		lines = append(lines[:unmanagedIndex], lines[unmanagedIndex+1:]...)
	}

	return networkManagerWriteConf(lines)
}

func networkManagerRemoveAllUnmanaged() error {
	lines, _, unmanagedIndex, err := networkManagerReadConf()
	if err != nil {
		return err
	}

	if unmanagedIndex == -1 {
		return nil
	}

	lines = append(lines[:unmanagedIndex], lines[unmanagedIndex+1:]...)

	return networkManagerWriteConf(lines)
}

func networkManagerWaitUntilUnmanaged(ifname string) bool {
	var out bytes.Buffer

	for tm := time.Now(); time.Since(tm).Seconds() < 10; {
		cmd := exec.Command("nmcli", "-t", "-f", "DEVICE,STATE", "d")
		cmd.Stdout = &out
		cmd.Stderr = &out

		if err := cmd.Run(); err != nil {
			return false
		}

		for {
			line, err := out.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					break
				}
				return false
			}

			if line == ifname+":unmanaged\n" {
				time.Sleep(500 * time.Millisecond)
				return true
			}
		}

		time.Sleep(500 * time.Millisecond)
	}

	return false
}
