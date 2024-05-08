package ipaddr

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
)

type Address struct {
	Address   string
	Interface string
	Version   int
}

type IPAddr struct {
	logger *slog.Logger
}

type IPAddrConfig struct {
	Logger *slog.Logger
}

func New(config IPAddrConfig) *IPAddr {
	return &IPAddr{
		logger: config.Logger,
	}
}

func (i *IPAddr) GetAddress(ip int, external bool, iface string) (string, error) {
	if external {
		addr, err := i.GetExternalAddress(ip)
		if err != nil {
			return "", err
		}

		return addr, nil
	} else {
		ifAddrs, err := i.GetIfAddresses()
		if err != nil {
			return "", err
		}

		for in, ifAddr := range ifAddrs {
			i.logger.Debug("GetIfAddresses()", "index", in, "address", ifAddr)
			if ifAddr.Version == ip && (iface == "" || ifAddr.Interface == iface) {
				return ifAddr.Address, nil
			}
		}

		return "", fmt.Errorf("Cannot get address of interface")
	}
}

func (i *IPAddr) GetIfAddresses() ([]*Address, error) {
	publicAddress := []*Address{}
	privateAddress := []*Address{}

	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	i.logger.Debug("net.Interfaces()", "interfaces", ifaces)

	for _, in := range ifaces {
		addrs, err := in.Addrs()
		if err != nil {
			return nil, err
		}
		i.logger.Debug("interfaces.Addrs()", "interface", in, "addresses", addrs)

		for _, a := range addrs {
			if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					ad := &Address{
						Address:   ipnet.IP.String(),
						Interface: in.Name,
						Version:   4,
					}
					if ipnet.IP.IsPrivate() {
						privateAddress = append(privateAddress, ad)
					} else {
						publicAddress = append(publicAddress, ad)
					}
				} else {
					ad := &Address{
						Address:   ipnet.IP.String(),
						Interface: in.Name,
						Version:   6,
					}
					if ipnet.IP.IsLinkLocalUnicast() {
						privateAddress = append(privateAddress, ad)
					} else {
						publicAddress = append(publicAddress, ad)
					}
				}
			}
		}
	}

	addresses := append(publicAddress, privateAddress...)

	return addresses, nil
}

func (i *IPAddr) GetExternalAddress(version int) (string, error) {
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
