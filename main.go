package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	Version = "0.3"
)

var (
	argExamples = kingpin.Command("examples", "Show examples for this tool.")

	argStart   = kingpin.Command("start", "Create new Access Point.")
	argGateway = argStart.Flag("gateway", "IPv4 Gateway for the AP.").
			Short('g').Default("192.168.12.1").String()
	argInterface = argStart.Flag("interface", "WiFi interface that will create the AP.").
			Short('i').Required().String()
	argSSID       = argStart.Flag("ssid", "Name of the AP.").Short('s').Required().String()
	argPassphrase = argStart.Flag("passphrase", "Set passphrase.").Short('p').String()
)

func main() {
	kingpin.Version(Version).Author("oblique")

	switch kingpin.Parse() {
	case "examples":
		cmdExamples()
	case "start":
		cmdStart()
	}
}

func cmdExamples() {
	printExamples()
}

func cmdStart() {
	var ap AccessPoint
	var err error

	// Check if all dependencies are installed
	if err = checkDeps(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if os.Geteuid() != 0 {
		fmt.Println("You must run it as root.")
		os.Exit(1)
	}

	if err = ap.wifiIf.init(*argInterface); err != nil {
		log.Fatalln(err)
	}

	ap.ssid = *argSSID
	if len(*argSSID) < 1 || len(*argSSID) > 32 {
		log.Fatalln("Invalid SSID length (expected 1..32)")
	}

	ap.passphrase = *argPassphrase
	if len(*argPassphrase) > 0 &&
		(len(*argPassphrase) < 8 || len(*argPassphrase) > 63) {
		log.Fatalln("Invalid passphrase length (expected 8..63)")
	}

	ap.wpa = WPA1 | WPA2
	ap.channel.num = 1
	ap.channel.mhz = 2412

	ap.gateway, err = parseIPv4(*argGateway)
	if err != nil {
		log.Fatalln(err)
	}

	if err = ap.start(); err != nil {
		ap.stop()
		log.Fatalln(err)
	}
	defer ap.stop()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGUSR1)

L:
	for {
		select {
		case <-sigs:
			log.Println("Exit signal received")
			break L
		case err := <-ap.fatalError:
			log.Println(err)
			break L
		default:
		}
		time.Sleep(100 * time.Millisecond)
	}

	log.Println("Exiting")
}
