package main

import "fmt"

func printExamples() {
	fmt.Print(`▶ Create Access Point
  • No passphrase (open network):
      create_ap start -s MyAccessPoint wlan0

  • WPA + WPA2 passphrase:
      create_ap start -s MyAccessPoint -p MyPassPhrase wlan0
`)
}
