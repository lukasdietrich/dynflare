package nameserver

import (
	"context"
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

func (c *cloudflareNameserver) lookupRecord(zoneId, domainName string, kind RecordKind) (*cloudflare.DNSRecord, error) {
	filter := cloudflare.DNSRecord{
		Type: string(kind),
		Name: domainName,
	}

	records, err := c.client.DNSRecords(context.Background(), zoneId, filter)
	if err != nil {
		return nil, fmt.Errorf("could not lookup records: %w", err)
	}

	if len(records) > 0 {
		return &records[0], nil
	}

	return nil, nil
}

func (c *cloudflareNameserver) UpdateRecord(record Record) error {
	zoneId, err := c.lookupZone(record.Zone)
	if err != nil {
		return err
	}

	dnsRecord, err := c.lookupRecord(zoneId, record.Domain, record.Kind)
	if err != nil {
		return err
	}

	if dnsRecord != nil {
		newRecord := *dnsRecord
		newRecord.Content = record.IP.String()

		if dnsRecord.Content != newRecord.Content {
			log.Debug().
				Str("id", newRecord.ID).
				Str("content", newRecord.Content).
				Msg("updating record")

			return c.client.UpdateDNSRecord(context.Background(), zoneId, newRecord.ID, newRecord)
		} else {
			log.Debug().
				Str("id", newRecord.ID).
				Str("content", newRecord.Content).
				Msg("record already up to date")
		}

		return nil
	}

	newRecord := cloudflare.DNSRecord{
		Type:    string(record.Kind),
		Name:    record.Domain,
		Content: record.IP.String(),
	}

	log.Debug().
		Str("domain", newRecord.Name).
		Str("type", newRecord.Type).
		Str("content", newRecord.Content).
		Msg("creating new record")

	_, err = c.client.CreateDNSRecord(context.Background(), zoneId, newRecord)
	return err
}
