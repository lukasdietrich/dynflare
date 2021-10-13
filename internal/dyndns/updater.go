package dyndns

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cloudflare/cloudflare-go"

	"github.com/lukasdietrich/dynflare/internal/config"
	"github.com/lukasdietrich/dynflare/internal/resolve"
)

type updater struct {
	client  *cloudflare.API
	records map[string][]cloudflare.DNSRecord
	zones   map[string]string
	cache   *cache
}

func newUpdater(token string) (*updater, error) {
	client, err := cloudflare.NewWithAPIToken(token)
	if err != nil {
		return nil, fmt.Errorf("could not create cloudflare client: %w", err)
	}

	cache, err := newCache()
	if err != nil {
		return nil, fmt.Errorf("could not open cache: %w", err)
	}

	return &updater{
		client:  client,
		records: make(map[string][]cloudflare.DNSRecord),
		zones:   make(map[string]string),
		cache:   cache,
	}, nil
}

func (u *updater) fetchDNSRecord(domain config.Domain) (*cloudflare.DNSRecord, error) {
	if u.records[domain.Zone] == nil {
		zoneId, err := u.client.ZoneIDByName(domain.Zone)
		if err != nil {
			return nil, err
		}

		records, err := u.client.DNSRecords(context.Background(), zoneId, cloudflare.DNSRecord{})
		if err != nil {
			return nil, err
		}

		u.zones[domain.Zone] = zoneId
		u.records[domain.Zone] = records
	}

	for _, record := range u.records[domain.Zone] {
		if record.Name == domain.Name && config.DomainKind(record.Type) == domain.Kind {
			return &record, nil
		}
	}

	return nil, nil
}

func (u *updater) updateDNSRecord(domain config.Domain, addr string) error {
	log.Printf("updating %s %s -> %s", domain.Name, domain.Kind, addr)

	record, err := u.fetchDNSRecord(domain)
	if err != nil {
		return err
	}

	defer u.cache.write(domain, addr)

	now := time.Now()

	if record != nil {
		r := *record
		r.ModifiedOn = now
		r.Content = addr

		log.Printf("updating existing record zoneName=%s, zoneId=%s, recordId=%s",
			r.ZoneName, r.ZoneID, r.ID)

		return u.client.UpdateDNSRecord(context.Background(), r.ZoneID, r.ID, r)
	} else {
		r := cloudflare.DNSRecord{
			Type:       string(domain.Kind),
			Name:       domain.Name,
			Content:    addr,
			CreatedOn:  now,
			ModifiedOn: now,
			ZoneID:     u.zones[domain.Zone],
			ZoneName:   domain.Zone,
		}

		log.Printf("creating new record zoneName=%s, zoneId=%s", r.ZoneName, r.ZoneID)

		_, err := u.client.CreateDNSRecord(context.Background(), r.ZoneID, r)
		return err
	}
}

func (u *updater) lastAddr(domain config.Domain) (string, error) {
	return u.cache.read(domain)
}

func (u *updater) currentAddr(domain config.Domain) (string, error) {
	ip, err := resolve.Resolve(domain)
	return ip.String(), err
}

func Update(cfg config.Config) error {
	updater, err := newUpdater(cfg.Cloudflare.Token)
	if err != nil {
		return fmt.Errorf("could not create updater: %w", err)
	}

	for _, domain := range cfg.Domains {
		lastAddr, err := updater.lastAddr(domain)
		if err != nil {
			return err
		}

		currentAddr, err := updater.currentAddr(domain)
		if err != nil {
			return err
		}

		if lastAddr != currentAddr {
			updater.updateDNSRecord(domain, currentAddr)
		}

	}

	return nil
}
