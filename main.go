package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/cloudflare/cloudflare-go"
	"github.com/jessevdk/go-flags"
	"golang.org/x/exp/slices"
)

type options struct {
	Dryrun    bool   `short:"d" long:"dry-run" description:"Don't create or update DNS record"`
	External  bool   `short:"e" long:"external" description:"Use external address instead of interface address"`
	Hostname  string `short:"n" long:"hostname" description:"Name to use instead of hostname"`
	Interface string `short:"i" long:"interface" description:"Interface to use address"`
	IPv4      bool   `short:"4" long:"ipv4" description:"Create or update only A record"`
	IPv6      bool   `short:"6" long:"ipv6" description:"Create or update only AAAA record"`
	Prefix    string `short:"p" long:"prefix" description:"Prefix of hostname"`
	Subdomain string `short:"s" long:"subdomain" description:"Subdomain name"`
	Args      struct {
		ZONE_NAME string
	} ` positional-args:"yes" required:"1"`
}

func main() {
	var opts options

	if _, err := flags.Parse(&opts); err != nil {
		flagsErr := err.(*flags.Error)
		if flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		}
		os.Exit(2)
	}

	apiToken := os.Getenv("CLOUDFLARE_API_TOKEN")
	if apiToken == "" {
		fmt.Fprintln(os.Stderr, "CLOUDFLARE_API_TOKEN is required")
		os.Exit(1)
	}

	if opts.Dryrun {
		fmt.Println("Dry-run")
	}

	if opts.Hostname == "" {
		hostname, err := os.Hostname()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to get hostname")
			os.Exit(1)
		}

		opts.Hostname = hostname
	}

	recordNames := []string{opts.Hostname, opts.Args.ZONE_NAME}

	if opts.Prefix != "" {
		recordNames = slices.Insert(recordNames, 0, opts.Prefix)
	}

	if opts.Subdomain != "" {
		recordNames = slices.Insert(recordNames, len(recordNames)-1, opts.Subdomain)
	}

	recordName := strings.Join(recordNames, ".")

	api, err := cloudflare.NewWithAPIToken(apiToken)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to login to Cloudflare:", err)
		os.Exit(1)
	}

	ctx := context.Background()

	zone, err := GetZoneFromName(api, ctx, opts.Args.ZONE_NAME)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to get a zone:", err)
		os.Exit(1)
	}

	family := []int{}
	if opts.IPv4 {
		family = append(family, 4)
	}
	if opts.IPv6 {
		family = append(family, 6)
	}
	if len(family) < 1 {
		family = append(family, 4, 6)
	}

	for _, ip := range family {
		addr, err := getAddr(ip, opts.External, opts.Interface)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to get an address:", err)
			os.Exit(1)
		}

		recordType := "A"
		if ip == 6 {
			recordType = "AAAA"
		}

		record, err := GetDNSRecordFromNameAndType(api, ctx, zone.ID, recordName, recordType)
		if err != nil && !errors.Is(err, DNSRecordNotFoundErr) {
			fmt.Fprintln(os.Stderr, "Failed to get a DNS record:", err)
			os.Exit(1)
		}

		switch {
		case errors.Is(err, DNSRecordNotFoundErr):
			{
				record = &cloudflare.DNSRecord{
					Type:    recordType,
					Name:    recordName,
					Content: addr,
					ZoneID:  zone.ID,
				}

				if !opts.Dryrun {
					resp, err := api.CreateDNSRecord(ctx, zone.ID, *record)
					if err != nil {
						fmt.Fprintln(os.Stderr, "Failed to create a DNS record:", err)
						os.Exit(1)
					}
					record = &resp.Result
				}

				fmt.Printf("CREATED: %s %s: %s\n", record.Name, record.Type, record.Content)
			}
		case addr == record.Content:
			{
				fmt.Printf("UNCHANGED: %s %s: %s\n", record.Name, record.Type, record.Content)
			}
		default:
			{
				oldAddr := record.Content

				record.Content = addr

				if !opts.Dryrun {
					err := api.UpdateDNSRecord(ctx, record.ZoneID, record.ID, *record)
					if err != nil {
						fmt.Fprintln(os.Stderr, "Failed to update a DNS record:", err)
						os.Exit(1)
					}
				}

				fmt.Printf("UPDATED: %s %s: %s => %s\n", record.Name, record.Type, oldAddr, record.Content)
			}
		}
	}
}

func getAddr(ip int, external bool, iface string) (string, error) {
	if external {
		addr, err := GetExternalAddress(ip)
		if err != nil {
			return "", err
		}

		return addr, nil
	} else {
		ifAddrs, err := GetIfAddresses()
		if err != nil {
			return "", err
		}

		for _, ifAddr := range ifAddrs {
			if ifAddr.Version == ip && (iface == "" || ifAddr.Interface == iface) {
				return ifAddr.Address, nil
			}
		}

		return "", fmt.Errorf("Cannot get address of interface")
	}
}
