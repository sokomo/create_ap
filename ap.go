package main

import "os/exec"

type WpaVersion uint

const (
	WPA1 WpaVersion = 1 << iota
	WPA2            = 1 << iota
)

type IEEE80211 uint

const (
	IEEE80211_G IEEE80211 = iota
	IEEE80211_N
)

type AccessPoint struct {
	ssid           string
	passphrase     string
	gateway        *IPNet
	channel        WifiChannel
	ieee80211      IEEE80211
	countryCode    string
	hiddenSSID     bool
	isolateClients bool
	wpa            WpaVersion
	wifiIf         WifiInterface
	daemons        []*exec.Cmd
	confDir        string
	fatalError     chan error
}
