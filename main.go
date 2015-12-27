package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

const (
	Version = "0.3"
)

func init() {
	// Make sure that all command outputs are in english, so we can parse them
	// correctly.
	if err := os.Setenv("LC_ALL", "C"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: create_ap [options] [<wifi-interface>"+
			" <internet-interface> <acess-point-name> [<passphrase>]]\n")
		flag.PrintDefaults()
	}
}

func readSSID() string {
	if len(os.Args) > 3 {
		return os.Args[3]
	}
	return ""
}

func readPassphrase() string {
	if len(os.Args) > 4 {
		return os.Args[4]
	}
	return ""
}

func main() {
	var ap AccessPoint

	flag.Parse()

	if len(os.Args) == 1 {
		flag.Usage()
		os.Exit(1)
	}

	// Check if all dependencies are installed
	if err := checkDeps(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := ap.wifiIf.init(os.Args[1]); err != nil {
		log.Fatalln(err)
	}

	if err := ap.internetIf.init(os.Args[2]); err != nil {
		log.Fatalln(err)
	}

	ap.ssid = readSSID()
	ap.passphrase = readPassphrase()

	ap.start()
}
