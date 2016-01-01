package main

import "fmt"

func printExamples() {
	fmt.Print(`▶ Create Access Point
  • No passphrase (open network)
      create_ap start -s MyAccessPoint wlan0

  • WPA + WPA2 passphrase
      create_ap start -s MyAccessPoint -p MyPassPhrase wlan0

  • Select specific channel (e.g. channel 6)
      create_ap start -c 6 -s MyAccessPoint -p MyPassPhrase wlan0

  • WPA2 passphrase
      create_ap start --wpa 2 -s MyAccessPoint -p MyPassPhrase wlan0

  • Enable IEEE 802.11ac
      create_ap start --80211 ac -s MyAccessPoint -p MyPassPhrase wlan0

  • Hidden AP
      create_ap start --hidden -s MyAccessPoint -p MyPassPhrase wlan0

  • Disable communication between clients
      create_ap start --isolate-clients -s MyAccessPoint -p MyPassPhrase wlan0
`)
}
