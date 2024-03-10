package cf

import (
	"context"
	"errors"

	"github.com/cloudflare/cloudflare-go"
)

var (
	DNSRecordNotFoundErr = errors.New("DNSRecord not found")
)

type DNSRecord cloudflare.DNSRecord

type CF struct {
	api *cloudflare.API
}

type Config struct {
	APIToken string
}

func New(config Config) (*CF, error) {
	api, err := cloudflare.NewWithAPIToken(config.APIToken)
	if err != nil {
		return nil, err
	}

	return &CF{api: api}, nil
}

func (c *CF) GetZoneIdByZoneName(ctx context.Context, zoneName string) (string, error) {
	zoneID, err := c.api.ZoneIDByName(zoneName)
	if err != nil {
		return "", err
	}

	return zoneID, nil
}

func (c *CF) GetDNSRecordFromNameAndType(ctx context.Context, zoneID string, name string, recordType string) (*DNSRecord, error) {
	records, _, err := c.api.ListDNSRecords(ctx, cloudflare.ZoneIdentifier(zoneID), cloudflare.ListDNSRecordsParams{Name: name, Type: recordType})
	if err != nil {
		panic(err)
	}

	if len(records) < 1 {
		return nil, DNSRecordNotFoundErr
	}

	return (*DNSRecord)(&records[0]), nil
}

func (c *CF) CreateDNSRecord(ctx context.Context, zoneID string, record *DNSRecord) (*DNSRecord, error) {
	res, err := c.api.CreateDNSRecord(ctx, cloudflare.ZoneIdentifier(zoneID), cloudflare.CreateDNSRecordParams{
		Type:    record.Type,
		Name:    record.Name,
		Content: record.Content,
		ZoneID:  zoneID,
	})
	if err != nil {
		return nil, err
	}

	return (*DNSRecord)(&res), nil
}

func (c *CF) UpdateDNSRecord(ctx context.Context, zoneID string, record *DNSRecord) (*DNSRecord, error) {
	res, err := c.api.UpdateDNSRecord(ctx, cloudflare.ZoneIdentifier(zoneID), cloudflare.UpdateDNSRecordParams{
		Content: record.Content,
		ID:      record.ID,
		Name:    record.Name,
		Type:    record.Type,
	})
	if err != nil {
		return nil, err
	}

	return (*DNSRecord)(&res), nil
}
