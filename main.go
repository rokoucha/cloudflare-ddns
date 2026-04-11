package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/rokoucha/cloudflare-ddns/cf"
	"github.com/rokoucha/cloudflare-ddns/ipaddr"
	"golang.org/x/exp/slices"
)

type options struct {
	Address   string `short:"a" long:"address" description:"Explicit IP address to register (overrides interface/external detection)" value-name:"ADDRESS"`
	Dryrun    bool   `short:"d" long:"dry-run" description:"Don't create or update DNS record"`
	External  bool   `short:"e" long:"external" description:"Use external address instead of interface address"`
	Hostname  string `short:"n" long:"hostname" description:"Name to use instead of hostname"`
	Interface string `short:"i" long:"interface" description:"Interface to use address"`
	IPv4      bool   `short:"4" long:"ipv4" description:"Create or update only A record"`
	IPv6      bool   `short:"6" long:"ipv6" description:"Create or update only AAAA record"`
	Prefix    string `short:"p" long:"prefix" description:"Prefix of hostname"`
	Suffix    string `short:"s" long:"suffix" description:"Suffix of hostname"`
	Verbose   bool   `short:"v" long:"verbose" description:"Show verbose debug information"`
	Watch     string `short:"w" long:"watch" description:"Run continuously, updating every INTERVAL (e.g. 5m, 1h)" value-name:"INTERVAL"`
	Args      struct {
		ZONE_NAME string
	} ` positional-args:"yes" required:"1"`
}

func syncOnce(
	ctx context.Context,
	logger *slog.Logger,
	c *cf.CF,
	i *ipaddr.IPAddr,
	opts options,
	zoneID string,
	recordName string,
	family []int,
) error {
	for _, ip := range family {
		var addr string
		if opts.Address != "" {
			addr = opts.Address
		} else {
			var err error
			addr, err = i.GetAddress(ip, opts.External, opts.Interface)
			if err != nil {
				return fmt.Errorf("failed to get address: %w", err)
			}
		}

		recordType := "A"
		if ip == 6 {
			recordType = "AAAA"
		}

		record, err := c.GetDNSRecordFromNameAndType(ctx, zoneID, recordName, recordType)
		if err != nil && !errors.Is(err, cf.DNSRecordNotFoundErr) {
			return fmt.Errorf("failed to get DNS record: %w", err)
		}

		switch {
		case errors.Is(err, cf.DNSRecordNotFoundErr):
			record = &cf.DNSRecord{
				Type:    recordType,
				Name:    recordName,
				Content: addr,
			}

			if !opts.Dryrun {
				record, err = c.CreateDNSRecord(ctx, zoneID, record)
				if err != nil {
					return fmt.Errorf("failed to create DNS record: %w", err)
				}
			}

			logger.Info("CREATED", "name", record.Name, "type", record.Type, "content", record.Content)

		case addr == record.Content:
			logger.Info("UNCHANGED", "name", record.Name, "type", record.Type, "content", record.Content)

		default:
			oldAddr := record.Content
			record.Content = addr

			if !opts.Dryrun {
				record, err = c.UpdateDNSRecord(ctx, zoneID, record)
				if err != nil {
					return fmt.Errorf("failed to update DNS record: %w", err)
				}
			}

			logger.Info("UPDATED", "name", record.Name, "type", record.Type, "old", oldAddr, "new", record.Content)
		}
	}

	return nil
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

	level := slog.LevelInfo
	if opts.Verbose {
		level = slog.LevelDebug
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))

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

	if opts.Address != "" {
		ip := net.ParseIP(opts.Address)
		if ip == nil {
			logger.Error("Invalid --address", "value", opts.Address)
			os.Exit(1)
		}
		addrFamily := 6
		if ip.To4() != nil {
			addrFamily = 4
		}
		if (addrFamily == 4 && opts.IPv6) || (addrFamily == 6 && opts.IPv4) {
			logger.Warn("--address family does not match --ipv4/--ipv6 filter, nothing to do")
			os.Exit(0)
		}
		family = []int{addrFamily}
	}

	i := ipaddr.New(ipaddr.IPAddrConfig{
		Logger: logger,
	})

	if opts.Watch == "" {
		if err := syncOnce(ctx, logger, c, i, opts, zoneID, recordName, family); err != nil {
			logger.Error("Sync failed", "err", err)
			os.Exit(1)
		}
		return
	}

	interval, err := time.ParseDuration(opts.Watch)
	if err != nil || interval <= 0 {
		logger.Error("Invalid --watch interval", "value", opts.Watch)
		os.Exit(2)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	logger.Info("Watch mode started", "interval", interval)
	if err := syncOnce(ctx, logger, c, i, opts, zoneID, recordName, family); err != nil {
		logger.Error("Sync failed", "err", err)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("Shutting down")
			return
		case <-ticker.C:
			if err := syncOnce(ctx, logger, c, i, opts, zoneID, recordName, family); err != nil {
				logger.Error("Sync failed", "err", err)
			}
		}
	}
}
