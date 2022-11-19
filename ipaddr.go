package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
)

type address struct {
	Address   string
	Interface string
	Version   int
}

func GetIfAddresses() ([]address, error) {
	addresses := []address{}

	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			return nil, err
		}

		for _, a := range addrs {
			if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					addresses = append(addresses, address{
						Address:   ipnet.IP.String(),
						Interface: i.Name,
						Version:   4,
					})
				} else {
					addresses = append(addresses, address{
						Address:   ipnet.IP.String(),
						Interface: i.Name,
						Version:   6,
					})
				}
			}
		}
	}

	return addresses, nil
}

func GetExternalAddress(version int) (string, error) {
	if version != 4 && version != 6 {
		return "", fmt.Errorf("Invalid IP version: %d", version)
	}

	url := fmt.Sprintf("https://v%d.ident.me", version)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
