package main

import (
	"context"
	"fmt"

	"github.com/cloudflare/cloudflare-go"
)

var (
	DNSRecordNotFoundErr = fmt.Errorf("DNSRecord not found")
	ZoneNotFoundErr      = fmt.Errorf("Zone not found")
)

func GetZoneFromName(api *cloudflare.API, ctx context.Context, name string) (*cloudflare.Zone, error) {
	zones, err := api.ListZones(ctx, name)
	if err != nil {
		return nil, err
	}

	if len(zones) < 1 {
		return nil, ZoneNotFoundErr
	}

	return &zones[0], nil
}

func GetDNSRecordFromNameAndType(api *cloudflare.API, ctx context.Context, zoneID string, name string, recordType string) (*cloudflare.DNSRecord, error) {
	records, err := api.DNSRecords(ctx, zoneID, cloudflare.DNSRecord{Name: name, Type: recordType})
	if err != nil {
		panic(err)
	}

	if len(records) < 1 {
		return nil, DNSRecordNotFoundErr
	}

	return &records[0], nil
}
