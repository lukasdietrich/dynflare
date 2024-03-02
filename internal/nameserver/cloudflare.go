package nameserver

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/cloudflare/cloudflare-go"
	"github.com/rs/zerolog/log"
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

func (c *cloudflareNameserver) lookupRecord(resource *cloudflare.ResourceContainer, domainName string, kind RecordKind) (*cloudflare.DNSRecord, error) {
	filter := cloudflare.ListDNSRecordsParams{
		Type: string(kind),
		Name: domainName,
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

func (c *cloudflareNameserver) UpdateRecord(record Record) error {
	err := c.updateRecord(record)

	if err != nil {
		var apiError *cloudflare.Error
		if errors.As(err, &apiError) && apiError.ClientError() {
			return wrapPermanentClientError(err)
		}
	}

	return err
}

func (c *cloudflareNameserver) updateRecord(record Record) error {
	zoneId, err := c.lookupZone(record.Zone)
	if err != nil {
		return err
	}

	resource := cloudflare.ZoneIdentifier(zoneId)

	dnsRecord, err := c.lookupRecord(resource, record.Domain, record.Kind)
	if err != nil {
		return err
	}

	if dnsRecord != nil {
		return c.updateExistingRecord(resource, record, dnsRecord)
	}

	return err
}

func (c *cloudflareNameserver) updateExistingRecord(resource *cloudflare.ResourceContainer, record Record, oldDnsRecord *cloudflare.DNSRecord) error {
	if record.IP.String() == oldDnsRecord.Content {
		log.Debug().
			Str("id", oldDnsRecord.ID).
			Str("content", oldDnsRecord.Content).
			Msg("record already up to date")

		return nil
	}

	updateRecordParams := cloudflare.UpdateDNSRecordParams{
		ID:      oldDnsRecord.ID,
		Content: record.IP.String(),
	}

	newDnsRecord, err := c.client.UpdateDNSRecord(context.Background(), resource, updateRecordParams)
	if err != nil {
		return err
	}

	log.Debug().
		Str("id", newDnsRecord.ID).
		Str("content", newDnsRecord.Content).
		Msg("updating record")

	return nil
}

func (c *cloudflareNameserver) createNewRecord(resource *cloudflare.ResourceContainer, record Record) error {
	createRecordParams := cloudflare.CreateDNSRecordParams{
		Type:    string(record.Kind),
		Name:    record.Domain,
		Content: record.IP.String(),
	}

	newDnsRecord, err := c.client.CreateDNSRecord(context.Background(), resource, createRecordParams)
	if err != nil {
		return err
	}

	log.Debug().
		Str("domain", newDnsRecord.Name).
		Str("type", newDnsRecord.Type).
		Str("content", newDnsRecord.Content).
		Msg("creating new record")

	return nil
}
