package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"
)

type IEEE80211 uint

const (
	IEEE80211_A  IEEE80211 = (1 << iota)
	IEEE80211_G            = (1 << iota)
	IEEE80211_N            = (1 << iota)
	IEEE80211_AC           = (1 << iota)
)

type WifiChannel struct {
	num uint
	mhz uint
}

type WifiInterface struct {
	Interface
	hwmodes  IEEE80211
	channels []WifiChannel
	module   string
}

func (wifiIf *WifiInterface) init(ifname string) error {
	err := wifiIf.Interface.init(ifname)
	if err != nil {
		return err
	}

	if !wifiIf.isWifi() {
		return fmt.Errorf("'%s' is not a WiFi interface", wifiIf)
	}

	// get kernel module
	path, err := os.Readlink("/sys/class/net/" + wifiIf.Name + "/device/driver/module")
	if err == nil {
		wifiIf.module = path[strings.LastIndex(path, "/")+1:]
	}

	info, err := wifiIf.getIwInfo()
	if err != nil {
		return err
	}

	re := regexp.MustCompile("\\t\\* ([0-9]+) MHz \\[([0-9]+)\\] .*")
	mAll := re.FindAllStringSubmatch(info, -1)
	for _, x := range mAll {
		if strings.Contains(x[0], "no IR") || strings.Contains(x[0], "disabled") {
			continue
		}

		mhz, err := strconv.ParseUint(x[1], 10, 32)
		if err != nil {
			continue
		}

		num, err := strconv.ParseUint(x[2], 10, 32)
		if err != nil {
			continue
		}

		switch {
		case mhz > 2400 && mhz < 2500:
			wifiIf.hwmodes |= IEEE80211_G
		case mhz > 5000 && mhz < 6000:
			wifiIf.hwmodes |= IEEE80211_A
		}

		wifiIf.channels = append(wifiIf.channels, WifiChannel{uint(num), uint(mhz)})
	}

	re = regexp.MustCompile("\\tCapabilities: 0x(.*)")
	m := re.FindStringSubmatch(info)
	if len(m) == 2 {
		cap, _ := strconv.ParseUint(m[1], 16, 16)
		// Threat any non-zero HT capabilities as N
		if cap > 0 {
			wifiIf.hwmodes |= IEEE80211_N
		}
	}

	if wifiIf.hwmodes&IEEE80211_A != 0 {
		re = regexp.MustCompile("\\tVHT Capabilities \\(0x(.*)\\):")
		m = re.FindStringSubmatch(info)
		if len(m) == 2 {
			cap, _ := strconv.ParseUint(m[1], 16, 32)
			// Threat any non-zero VHT capabilities as AC
			if cap > 0 {
				wifiIf.hwmodes |= IEEE80211_AC
			}
		}
	}

	return nil
}

func (wifiIf *WifiInterface) getPhy() (string, error) {
	const ieee80211_dir = "/sys/class/ieee80211"

	files, err := ioutil.ReadDir(ieee80211_dir)
	if err != nil {
		return "", err
	}

	for _, x := range files {
		phy := x.Name()

		if phy == wifiIf.Name {
			return phy, nil
		}

		p := path.Join(ieee80211_dir, phy, "device/net", wifiIf.Name)
		if _, err := os.Lstat(p); err == nil {
			return phy, nil
		}

		p = path.Join(ieee80211_dir, phy, "device/net:"+wifiIf.Name)
		if _, err := os.Lstat(p); err == nil {
			return phy, nil
		}
	}

	return "", fmt.Errorf("Failed to get phy interface.")
}

func (wifiIf *WifiInterface) getIwInfo() (string, error) {
	var out bytes.Buffer

	phy, err := wifiIf.getPhy()
	if err != nil {
		return "", err
	}

	cmd := exec.Command("iw", "phy", phy, "info")
	cmd.Stdout = &out
	cmd.Stderr = &out

	err = cmd.Run()
	if err != nil || out.Len() == 0 {
		return "", fmt.Errorf("Failed to get adapter info.")
	}

	return out.String(), nil
}

func (wifiIf *WifiInterface) canTransmitOnChannel(channel uint) bool {
	for _, c := range wifiIf.channels {
		if c.num == channel {
			return true
		}
	}
	return false
}

func (wifiIf *WifiInterface) canAutoSelectChannel() bool {
	var out bytes.Buffer

	// Adapter must support survey dump to automatically select channel
	cmd := exec.Command("iw", "dev", wifiIf.Name, "survey", "dump")
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	if err != nil || out.Len() == 0 {
		return false
	}

	return true
}
