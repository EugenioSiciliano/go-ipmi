package main

import (
	"fmt"
	"github.com/bougou/go-ipmi"
)

// test client program
func main() {
	host := "172.17.3.190"
	port := 623
	username := "root"
	password := "calvin"

	client, err := ipmi.NewClient(host, port, username, password)
	// Support local mode client if runs directly on linux
	// client, err := ipmi.NewOpenClient()
	if err != nil {
		panic(err)
	}

	client.WithInterface(ipmi.InterfaceLanplus)
	// you can optionally open debug switch
	// client.WithDebug(true)

	// Connect will create an authenticated session for you.
	if err := client.Connect(); err != nil {
		panic(err)
	}

	// Now you can execute other IPMI commands that need authentication.

	dellClient, err := ipmi.NewDellOEMClient(client)
	if err != nil {
		panic(err)
	}

	list, err := dellClient.GetDellOEMMACs()
	if err != nil {
		panic(err)
	}

	for _, mac := range list {
		fmt.Println(fmt.Sprintf("found interface %q", mac))
	}
}
