package main

type WpaVersion uint

const (
	WPA1 WpaVersion = 1 << iota
	WPA2 WpaVersion = 1 << iota
)

type AccessPoint struct {
	ssid       string
	passphrase string
	channel    WifiChannel
	wpa        WpaVersion
	wifiIf     WifiInterface
	internetIf Interface
}
