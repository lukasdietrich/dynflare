package nameserver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/cloudflare/cloudflare-go"
)

type cloudflareNameserver struct {
	client *cloudflare.API

	zoneLock  sync.Mutex
	zoneCache map[string]string // zoneName -> zoneId
}

func newCloudflare(token string) (Nameserver, error) {
	client, err := cloudflare.NewWithAPIToken(token)
	if err != nil {
		return nil, err
	}

	return &cloudflareNameserver{client: client}, nil
}

func (c *cloudflareNameserver) lookupZone(zoneName string) (string, error) {
	c.zoneLock.Lock()
	defer c.zoneLock.Unlock()

	if c.zoneCache == nil {
		c.zoneCache = make(map[string]string)
	}

	zoneId, ok := c.zoneCache[zoneName]
	if ok {
		return zoneId, nil
	}

	zoneId, err := c.client.ZoneIDByName(zoneName)
	if err != nil {
		return zoneId, fmt.Errorf("could not lookup zone: %w", err)
	}

	c.zoneCache[zoneName] = zoneId
	return zoneId, nil
}

func (c *cloudflareNameserver) lookupRecord(resource *cloudflare.ResourceContainer, record Record) (*cloudflare.DNSRecord, error) {
	filter := cloudflare.ListDNSRecordsParams{
		Type:    string(record.Kind),
		Name:    record.Domain,
		Comment: record.Comment,
	}

	records, _, err := c.client.ListDNSRecords(context.Background(), resource, filter)
	if err != nil {
		return nil, fmt.Errorf("could not lookup records: %w", err)
	}

	if len(records) > 0 {
		return &records[0], nil
	}

	return nil, nil
}

func (c *cloudflareNameserver) UpdateRecord(record Record) (bool, error) {
	changed, err := c.updateRecord(record)

	if err != nil {
		var apiError *cloudflare.Error
		if errors.As(err, &apiError) && apiError.ClientError() {
			return false, wrapPermanentClientError(err)
		}
	}

	return changed, err
}

func (c *cloudflareNameserver) updateRecord(record Record) (bool, error) {
	zoneId, err := c.lookupZone(record.Zone)
	if err != nil {
		return false, err
	}

	resource := cloudflare.ZoneIdentifier(zoneId)

	dnsRecord, err := c.lookupRecord(resource, record)
	if err != nil {
		return false, err
	}

	if dnsRecord != nil {
		return c.updateExistingRecord(resource, record, dnsRecord)
	}

	if err := c.createNewRecord(resource, record); err != nil {
		return false, err
	}

	return true, nil
}

func (c *cloudflareNameserver) updateExistingRecord(
	resource *cloudflare.ResourceContainer,
	record Record,
	oldDnsRecord *cloudflare.DNSRecord,
) (bool, error) {
	if record.IP.String() == oldDnsRecord.Content {
		slog.Debug("record already up to date",
			slog.String("id", oldDnsRecord.ID),
			slog.String("content", oldDnsRecord.Content))

		return false, nil
	}

	updateRecordParams := cloudflare.UpdateDNSRecordParams{
		ID:      oldDnsRecord.ID,
		Content: record.IP.String(),
		Comment: &record.Comment,
	}

	newDnsRecord, err := c.client.UpdateDNSRecord(context.Background(), resource, updateRecordParams)
	if err != nil {
		return false, err
	}

	slog.Debug("updating record",
		slog.String("id", newDnsRecord.ID),
		slog.String("content", newDnsRecord.Content))

	return true, nil
}

func (c *cloudflareNameserver) createNewRecord(resource *cloudflare.ResourceContainer, record Record) error {
	createRecordParams := cloudflare.CreateDNSRecordParams{
		Type:    string(record.Kind),
		Name:    record.Domain,
		Content: record.IP.String(),
		Comment: record.Comment,
	}

	newDnsRecord, err := c.client.CreateDNSRecord(context.Background(), resource, createRecordParams)
	if err != nil {
		return err
	}

	slog.Debug("creating new record",
		slog.String("domain", newDnsRecord.Name),
		slog.String("type", newDnsRecord.Type),
		slog.String("content", newDnsRecord.Content))

	return nil
}
