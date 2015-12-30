package main

import "fmt"

func printExamples() {
	fmt.Print(`▶ Create Access Point
  • No passphrase (open network):
      create_ap start -i wlan0 -s MyAccessPoint

  • WPA + WPA2 passphrase:
      create_ap start -i wlan0 -s MyAccessPoint -p MyPassPhrase
`)
}
