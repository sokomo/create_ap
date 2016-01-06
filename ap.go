package main

import "os/exec"

type WpaVersion uint

const (
	WPA1 WpaVersion = 1 << iota
	WPA2            = 1 << iota
)

type AccessPoint struct {
	ssid           string
	passphrase     string
	gateway        *IPNet
	channel        uint
	ieee80211      string
	countryCode    string
	hiddenSSID     bool
	isolateClients bool
	wpa            WpaVersion
	wifiIf         WifiInterface
	daemons        []*exec.Cmd
	confDir        string
	fatalError     chan error
}
