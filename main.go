package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/rokoucha/cloudflare-ddns/cf"
	"github.com/rokoucha/cloudflare-ddns/ipaddr"
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
	Suffix    string `short:"s" long:"suffix" description:"Suffix of hostname"`
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

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	apiToken := os.Getenv("CLOUDFLARE_API_TOKEN")
	if apiToken == "" {
		logger.Error("CLOUDFLARE_API_TOKEN is required")
		os.Exit(1)
	}

	if opts.Dryrun {
		logger.Info("Dry-run")
	}

	if opts.Hostname == "" {
		hostname, err := os.Hostname()
		if err != nil {
			logger.Error("Failed to get hostname", "err", err)
			os.Exit(1)
		}

		opts.Hostname = hostname
	}

	recordNames := []string{opts.Hostname, opts.Args.ZONE_NAME}

	if opts.Prefix != "" {
		recordNames = slices.Insert(recordNames, 0, opts.Prefix)
	}

	if opts.Suffix != "" {
		recordNames = slices.Insert(recordNames, len(recordNames)-1, opts.Suffix)
	}

	recordName := strings.Join(recordNames, ".")

	c, err := cf.New(cf.Config{APIToken: apiToken})
	if err != nil {
		logger.Error("Failed to login to cloudflare", "err", err)
		os.Exit(1)
	}

	ctx := context.Background()

	zoneID, err := c.GetZoneIdByZoneName(ctx, opts.Args.ZONE_NAME)
	if err != nil {
		logger.Error("Failed to get a zone", "err", err)
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
		addr, err := ipaddr.GetAddress(ip, opts.External, opts.Interface)
		if err != nil {
			logger.Error("Failed to get an address", "err", err)
			os.Exit(1)
		}

		recordType := "A"
		if ip == 6 {
			recordType = "AAAA"
		}

		record, err := c.GetDNSRecordFromNameAndType(ctx, zoneID, recordName, recordType)
		if err != nil && !errors.Is(err, cf.DNSRecordNotFoundErr) {
			logger.Error("Failed to get a DNS record", "err", err)
			os.Exit(1)
		}

		switch {
		case errors.Is(err, cf.DNSRecordNotFoundErr):
			{
				record = &cf.DNSRecord{
					Type:    recordType,
					Name:    recordName,
					Content: addr,
					ZoneID:  zoneID,
				}

				if !opts.Dryrun {
					record, err = c.CreateDNSRecord(ctx, zoneID, record)
					if err != nil {
						logger.Error("Failed to create a DNS record", "err", err)
						os.Exit(1)
					}
				}

				logger.Info("CREATED", "name", record.Name, "type", record.Type, "content", record.Content)
			}
		case addr == record.Content:
			{
				logger.Info("UNCHANGED", "name", record.Name, "type", record.Type, "content", record.Content)
			}
		default:
			{
				oldAddr := record.Content

				record.Content = addr

				if !opts.Dryrun {
					record, err = c.UpdateDNSRecord(ctx, zoneID, record)
					if err != nil {
						logger.Error("Failed to update a DNS record", "err", err)
						os.Exit(1)
					}
				}

				logger.Info("UPDATED", "name", record.Name, "type", record.Type, "old", oldAddr, "new", record.Content)
			}
		}
	}
}
