package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

type Interface struct {
	*net.Interface
}

func (iface *Interface) init(ifname string) error {
	ifi, err := net.InterfaceByName(ifname)
	if err != nil {
		return fmt.Errorf("%s (%s)", err, ifname)
	}

	iface.Interface = ifi
	return nil
}

func (iface *Interface) isWifi() bool {
	_, err := os.Stat("/sys/class/net/" + iface.Name + "/wireless")
	return err == nil
}

func (iface *Interface) isBridge() bool {
	_, err := os.Stat("/sys/class/net/" + iface.Name + "/bridge")
	return err == nil
}

type WifiChannel struct {
	num uint
	mhz uint
}

type WifiInterface struct {
	Interface
	module string
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

	return nil
}
