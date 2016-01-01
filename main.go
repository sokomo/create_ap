package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	Version = "1.0.0-alpha"
)

var (
	argExamples = kingpin.Command("examples", "Show examples for this tool.")

	argStart     = kingpin.Command("start", "Create new Access Point.")
	argInterface = argStart.Arg("interface", "WiFi interface that will create the AP.").
			Required().String()
	argGateway = argStart.Flag("gateway", "IPv4 Gateway for the AP.").
			Short('g').Default("192.168.12.1").String()
	argSSID           = argStart.Flag("ssid", "Name of the AP.").Short('s').Required().String()
	argPassphrase     = argStart.Flag("passphrase", "Set passphrase.").Short('p').String()
	argWPA            = argStart.Flag("wpa", "Set WPA versions").Default("1,2").String()
	argChannel        = argStart.Flag("channel", "Set channel number").Short('c').Default("1").Uint()
	argHidden         = argStart.Flag("hidden", "Make AP hidden (i.e. do not broadcast SSID)").Bool()
	argIsolateClients = argStart.Flag("isolate-clients", "Disable communication between clients").Bool()
	arg80211          = argStart.Flag("80211", "Set 802.11 protocol. Valid inputs: g, n, ac").
				Default("n").String()
	argCountry = argStart.Flag("country", "Set two-letter country code for regularity").
			Default("00").String()
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

func parseArgWPA() (WpaVersion, error) {
	var wpa WpaVersion

	for _, s := range strings.Split(*argWPA, ",") {
		switch s {
		case "1":
			wpa |= WPA1
		case "2":
			wpa |= WPA2
		case "":
		default:
			return 0, fmt.Errorf("Invalid WPA version (%s)", s)
		}
	}

	return wpa, nil
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

	// Setup configuration
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

	switch strings.ToLower(*arg80211) {
	case "g":
		ap.ieee80211 = IEEE80211_G
	case "n":
		ap.ieee80211 = IEEE80211_N
	case "ac":
		ap.ieee80211 = IEEE80211_AC
	default:
		log.Fatalln("Invalid 802.11 protocol")
	}

	ap.wpa, err = parseArgWPA()
	if err != nil {
		log.Fatalln(err)
	}

	ap.countryCode = strings.ToUpper(*argCountry)
	if len(ap.countryCode) != 2 {
		log.Fatalln("Invalid country code")
	}

	ap.channel = *argChannel
	ap.hiddenSSID = *argHidden
	ap.isolateClients = *argIsolateClients

	ap.gateway, err = parseIPv4(*argGateway)
	if err != nil {
		log.Fatalln(err)
	}

	// Start AP
	if err = ap.start(); err != nil {
		ap.stop()
		log.Fatalln(err)
	}
	defer ap.stop()

	// Wait for exit signal
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, os.Kill, syscall.SIGTERM)

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
