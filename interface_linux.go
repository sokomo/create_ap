package main

import (
	"fmt"
	"net"
	"os"
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

func (iface *Interface) refresh() error {
	return iface.init(iface.Name)
}

func (iface *Interface) setDown() error {
	err := runCmd("ip", "link", "set", "down", "dev", iface.Name)
	if err != nil {
		return err
	}

	err = runCmd("ip", "addr", "flush", iface.Name)
	if err != nil {
		return err
	}

	return iface.refresh()
}

func (iface *Interface) addIPv4(ipnet *IPNet) error {
	return runCmd("ip", "addr", "add", ipnet.String(),
		"broadcast", ipnet.broadcast().String(), "dev", iface.Name)
}
