package caddy_cfrl

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/cloudflare/cloudflare-go"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

func init() {
	caddy.RegisterModule(CloudflareRateLimit{})
}

// CloudflareRateLimit is a Caddy module that rate limits IPs and adds them to a Cloudflare IP list.
type CloudflareRateLimit struct {
	APIToken      string         `json:"api_token,omitempty"`
	AccountID     string         `json:"account_id,omitempty"`
	ListID        string         `json:"list_id,omitempty"`
	Rate          float64        `json:"rate,omitempty"`
	Burst         int            `json:"burst,omitempty"`
	EntryDuration caddy.Duration `json:"entry_duration,omitempty"`

	client   *cloudflare.API
	limiters sync.Map
	logger   *zap.Logger
}

// Provision sets up the module.
func (c *CloudflareRateLimit) Provision(ctx caddy.Context) error {
	c.logger = ctx.Logger(c)

	if c.APIToken == "" {
		return fmt.Errorf("missing cloudflare api_token")
	}
	if c.AccountID == "" {
		return fmt.Errorf("missing cloudflare account_id")
	}
	if c.ListID == "" {
		return fmt.Errorf("missing cloudflare list_id")
	}

	var err error
	c.client, err = cloudflare.NewWithAPIToken(c.APIToken)
	if err != nil {
		return fmt.Errorf("failed to create cloudflare client: %v", err)
	}

	return nil
}

// Validate ensures the configuration is valid.
func (c *CloudflareRateLimit) Validate() error {
	if c.Rate <= 0 {
		return fmt.Errorf("rate must be greater than 0")
	}
	if c.Burst <= 0 {
		return fmt.Errorf("burst must be greater than 0")
	}
	if c.EntryDuration <= 0 {
		return fmt.Errorf("entry_duration must be greater than 0")
	}
	return nil
}

// ServeHTTP implements caddyhttp.Handler.
func (c *CloudflareRateLimit) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		ip = r.RemoteAddr
	}

	limiterRaw, _ := c.limiters.LoadOrStore(ip, rate.NewLimiter(rate.Limit(c.Rate), c.Burst))
	limiter := limiterRaw.(*rate.Limiter)

	if !limiter.Allow() {
		c.logger.Warn("rate limit exceeded, adding to cloudflare",
			zap.String("ip", ip),
			zap.Duration("duration", time.Duration(c.EntryDuration)))

		go c.blockIP(ip)

		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte("Rate limit exceeded. Your IP has been temporarily blocked."))
		return nil
	}

	return next.ServeHTTP(w, r)
}

func (c *CloudflareRateLimit) blockIP(ip string) {
	ctx := context.Background()

	// Add IP to Cloudflare List
	_, err := c.client.CreateListItems(ctx, cloudflare.AccountIdentifier(c.AccountID), cloudflare.ListCreateItemsParams{
		ID: c.ListID,
		Items: []cloudflare.ListItemCreateRequest{
			{IP: &ip, Comment: "Caddy Ratelimit Block"},
		},
	})
	if err != nil {
		c.logger.Error("failed to add IP to cloudflare list",
			zap.String("ip", ip),
			zap.Error(err))
		return
	}

	// Schedule removal
	time.AfterFunc(time.Duration(c.EntryDuration), func() {
		err := c.unblockIP(ip)
		if err != nil {
			c.logger.Error("failed to remove IP from cloudflare list",
				zap.String("ip", ip),
				zap.Error(err))
		}
	})
}

func (c *CloudflareRateLimit) unblockIP(ip string) error {
	ctx := context.Background()

	// To delete an item, we first need to find its ID in the list
	items, err := c.client.ListListItems(ctx, cloudflare.AccountIdentifier(c.AccountID), cloudflare.ListListItemsParams{
		ID: c.ListID,
	})
	if err != nil {
		return fmt.Errorf("failed to list items: %v", err)
	}

	var itemID string
	for _, item := range items {
		if item.IP != nil && *item.IP == ip {
			itemID = item.ID
			break
		}
	}

	if itemID == "" {
		return fmt.Errorf("ip %s not found in list", ip)
	}

	_, err = c.client.DeleteListItems(ctx, cloudflare.AccountIdentifier(c.AccountID), cloudflare.ListDeleteItemsParams{
		ID: c.ListID,
		Items: cloudflare.ListItemDeleteRequest{
			Items: []cloudflare.ListItemDeleteItemRequest{
				{ID: itemID},
			},
		},
	})
	return err
}

// UnmarshalCaddyfile implements caddyfile.Unmarshaler.
func (c *CloudflareRateLimit) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		for d.NextBlock(0) {
			switch d.Val() {
			case "api_token":
				if !d.NextArg() {
					return d.ArgErr()
				}
				c.APIToken = d.Val()
			case "account_id":
				if !d.NextArg() {
					return d.ArgErr()
				}
				c.AccountID = d.Val()
			case "list_id":
				if !d.NextArg() {
					return d.ArgErr()
				}
				c.ListID = d.Val()
			case "rate":
				if !d.NextArg() {
					return d.ArgErr()
				}
				fmt.Sscanf(d.Val(), "%f", &c.Rate)
			case "burst":
				if !d.NextArg() {
					return d.ArgErr()
				}
				fmt.Sscanf(d.Val(), "%d", &c.Burst)
			case "entry_duration":
				if !d.NextArg() {
					return d.ArgErr()
				}
				dur, err := caddy.ParseDuration(d.Val())
				if err != nil {
					return d.Errf("invalid duration: %v", err)
				}
				c.EntryDuration = caddy.Duration(dur)
			}
		}
	}
	return nil
}

// Interface guards
var (
	_ caddy.Provisioner           = (*CloudflareRateLimit)(nil)
	_ caddy.Validator             = (*CloudflareRateLimit)(nil)
	_ caddyhttp.MiddlewareHandler = (*CloudflareRateLimit)(nil)
	_ caddyfile.Unmarshaler       = (*CloudflareRateLimit)(nil)
)

// CaddyModule returns the Caddy module information.
func (CloudflareRateLimit) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.cloudflare_ratelimit",
		New: func() caddy.Module { return new(CloudflareRateLimit) },
	}
}
