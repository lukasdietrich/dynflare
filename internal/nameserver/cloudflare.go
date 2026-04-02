package nameserver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/dns"
	"github.com/cloudflare/cloudflare-go/v6/option"
	"github.com/cloudflare/cloudflare-go/v6/zones"
)

type cloudflareNameserver struct {
	client *cloudflare.Client

	zoneLock  sync.Mutex
	zoneCache map[string]string // zoneName -> zoneId
}

func newCloudflare(token string) Nameserver {
	client := cloudflare.NewClient(option.WithAPIToken(token))
	return &cloudflareNameserver{client: client}
}

func (c *cloudflareNameserver) lookupZone(ctx context.Context, zoneName string) (string, error) {
	c.zoneLock.Lock()
	defer c.zoneLock.Unlock()

	if c.zoneCache == nil {
		c.zoneCache = make(map[string]string)
	}

	zoneId, ok := c.zoneCache[zoneName]
	if ok {
		return zoneId, nil
	}

	filter := zones.ZoneListParams{
		Match: cloudflare.F(zones.ZoneListParamsMatchAll),
		Name:  cloudflare.F(zoneName),
	}

	zonePagination, err := c.client.Zones.List(ctx, filter)
	if err != nil {
		return zoneId, fmt.Errorf("could not lookup zone: %w", err)
	}

	if len(zonePagination.Result) != 1 {
		return zoneId, fmt.Errorf("got %d zones for name %q", len(zonePagination.Result), zoneName)
	}

	zoneId = zonePagination.Result[0].ID
	c.zoneCache[zoneName] = zoneId
	return zoneId, nil
}

func (c *cloudflareNameserver) lookupRecord(ctx context.Context, zoneId string, record Record) (*dns.RecordResponse, error) {
	filter := dns.RecordListParams{
		Match:  cloudflare.F(dns.RecordListParamsMatchAll),
		Type:   cloudflare.F(dns.RecordListParamsType(record.Kind)),
		ZoneID: cloudflare.F(zoneId),
		Name: cloudflare.F(dns.RecordListParamsName{
			Exact: cloudflare.F(record.Domain),
		}),
	}

	if record.Comment != "" {
		filter.Comment = cloudflare.F(dns.RecordListParamsComment{
			Exact: cloudflare.F(record.Comment),
		})
	}

	recordPagination, err := c.client.DNS.Records.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("could not lookup records: %w", err)
	}

	if len(recordPagination.Result) > 1 {
		return nil, fmt.Errorf("got %d records for name %q", len(recordPagination.Result), record.Domain)
	}

	if len(recordPagination.Result) == 0 {
		return nil, nil
	}

	return &recordPagination.Result[0], nil
}

func (c *cloudflareNameserver) UpdateRecord(ctx context.Context, record Record) (bool, error) {
	changed, err := c.updateRecord(ctx, record)

	if err != nil {
		var apiError *cloudflare.Error
		if errors.As(err, &apiError) && isClientError(apiError.StatusCode) {
			return false, wrapPermanentClientError(err)
		}
	}

	return changed, err
}

func isClientError(status int) bool {
	return status >= 400 && status < 500
}

func (c *cloudflareNameserver) updateRecord(ctx context.Context, record Record) (bool, error) {
	zoneId, err := c.lookupZone(ctx, record.Zone)
	if err != nil {
		return false, err
	}

	dnsRecord, err := c.lookupRecord(ctx, zoneId, record)
	if err != nil {
		return false, err
	}

	if dnsRecord != nil {
		return c.updateExistingRecord(ctx, zoneId, record, dnsRecord)
	}

	if err := c.createNewRecord(ctx, zoneId, record); err != nil {
		return false, err
	}

	return true, nil
}

func (c *cloudflareNameserver) updateExistingRecord(ctx context.Context, zoneId string, record Record, oldDnsRecord *dns.RecordResponse) (bool, error) {
	if record.IP.String() == oldDnsRecord.Content {
		slog.Debug("record already up to date",
			slog.String("id", oldDnsRecord.ID),
			slog.String("content", oldDnsRecord.Content))

		return false, nil
	}

	updateRecordParams := dns.RecordUpdateParams{
		ZoneID: cloudflare.F(zoneId),
		Body: dns.RecordUpdateParamsBody{
			Type:    cloudflare.F(dns.RecordUpdateParamsBodyType(record.Kind)),
			Name:    cloudflare.F(record.Domain),
			Comment: cloudflare.F(record.Comment),
			Content: cloudflare.F(record.IP.String()),
		},
	}

	newDnsRecord, err := c.client.DNS.Records.Update(ctx, oldDnsRecord.ID, updateRecordParams)
	if err != nil {
		return false, err
	}

	slog.Debug("updating record",
		slog.String("id", newDnsRecord.ID),
		slog.String("content", newDnsRecord.Content))

	return true, nil
}

func (c *cloudflareNameserver) createNewRecord(ctx context.Context, zoneId string, record Record) error {
	createRecordParams := dns.RecordNewParams{
		ZoneID: cloudflare.F(zoneId),
		Body: dns.RecordNewParamsBody{
			Type:    cloudflare.F(dns.RecordNewParamsBodyType(record.Kind)),
			Name:    cloudflare.F(record.Domain),
			Comment: cloudflare.F(record.Comment),
			Content: cloudflare.F(record.IP.String()),
		},
	}

	newDnsRecord, err := c.client.DNS.Records.New(ctx, createRecordParams)
	if err != nil {
		return err
	}

	slog.Debug("creating new record",
		slog.String("domain", newDnsRecord.Name),
		slog.String("type", string(newDnsRecord.Type)),
		slog.String("content", newDnsRecord.Content))

	return nil
}
