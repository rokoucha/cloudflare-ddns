package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/cloudflare/cloudflare-go"
	"github.com/jessevdk/go-flags"
	"golang.org/x/exp/slices"
)

type options struct {
	Dryrun    bool   `short:"d" long:"dry-run" description:"dry run"`
	External  bool   `short:"e" long:"external" description:"Use external global address"`
	Hostname  string `short:"n" long:"hostname" description:"hostname"`
	Interface string `short:"i" long:"interface" description:"interface name"`
	IPv4      bool   `short:"4" long:"ipv4" description:"Add A record"`
	IPv6      bool   `short:"6" long:"ipv6" description:"Add AAAA record"`
	Prefix    string `short:"p" long:"prefix" description:"Prefix"`
	Subdomain string `short:"s" long:"subdomain" description:"Subdomain"`
}

func main() {
	var opts options

	args, err := flags.Parse(&opts)
	if err != nil {
		flagsErr := err.(*flags.Error)
		if flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		}
		os.Exit(2)
	}

	if len(args) != 1 || args[0] == "" {
		log.Fatalln("missing operand")
	}

	zoneName := args[0]

	apiToken := os.Getenv("CLOUDFLARE_API_TOKEN")
	if apiToken == "" {
		log.Fatalln("CLOUDFLARE_API_TOKEN is required")
	}

	if opts.Dryrun {
		log.Println("Dry-run")
	}

	if opts.Hostname == "" {
		hostname, err := os.Hostname()
		if err != nil {
			log.Fatalln(err)
		}

		opts.Hostname = hostname
	}

	recordNames := []string{opts.Hostname, zoneName}

	if opts.Prefix != "" {
		recordNames = slices.Insert(recordNames, 0, opts.Prefix)
	}

	if opts.Subdomain != "" {
		recordNames = slices.Insert(recordNames, len(recordNames)-1, opts.Subdomain)
	}

	recordName := strings.Join(recordNames, ".")

	api, err := cloudflare.NewWithAPIToken(apiToken)
	if err != nil {
		log.Fatalln(err)
	}

	ctx := context.Background()

	zone, err := GetZoneFromName(api, ctx, zoneName)
	if err != nil {
		log.Fatalln(err)
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
			log.Fatalln(err)
		}

		recordType := "A"
		if ip == 6 {
			recordType = "AAAA"
		}

		record, err := GetDNSRecordFromNameAndType(api, ctx, zone.ID, recordName, recordType)
		if err != nil && !errors.Is(err, DNSRecordNotFoundErr) {
			log.Fatalln(err)
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
						log.Fatalln(err)
					}
					record = &resp.Result
				}

				log.Printf("CREATED: %s %s: %s\n", record.Name, record.Type, record.Content)
			}
		case addr == record.Content:
			{
				log.Printf("UNCHANGED: %s %s: %s\n", record.Name, record.Type, record.Content)
			}
		default:
			{
				oldAddr := record.Content

				record.Content = addr

				if !opts.Dryrun {
					err := api.UpdateDNSRecord(ctx, record.ZoneID, record.ID, *record)
					if err != nil {
						log.Fatalln(err)
					}
				}

				log.Printf("UPDATED: %s %s: %s => %s\n", record.Name, record.Type, oldAddr, record.Content)
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
