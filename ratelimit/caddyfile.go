package ratelimit

import (
	"strconv"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

func init() {
	httpcaddyfile.RegisterHandlerDirective("rate_limit", parseCaddyfile)
}

// parseCaddyfile sets up a handler for rate-limiting from Caddyfile tokens. Syntax:
//
//	rate_limit [<matcher>] <key> <rate> [<zone_size> [<reject_status>]]
//
// Parameters:
// - <key>: The variable used to differentiate one client from another.
// - <rate>: The request rate limit (per key value) specified in requests per second (r/s) or requests per minute (r/m).
// - <zone_size>: The size (i.e. the number of key values) of the LRU zone that keeps states of these key values. Defaults to 10,000.
// - <reject_status>: The HTTP status code of the response when a client exceeds the rate. Defaults to 429 (Too Many Requests).
func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	rl := new(RateLimit)
	if err := rl.UnmarshalCaddyfile(h.Dispenser); err != nil {
		return nil, err
	}
	return rl, nil
}

func (rl *RateLimit) UnmarshalCaddyfile(d *caddyfile.Dispenser) (err error) {
	for d.Next() {
		args := d.RemainingArgs()
		switch len(args) {
		case 4:
			rl.RejectStatusCode, err = strconv.Atoi(args[3])
			if err != nil {
				return d.Errf("reject_status must be an integer; invalid: %v", err)
			}
			fallthrough
		case 3:
			size, err := strconv.Atoi(args[2])
			if err != nil {
				return d.Errf("zone_size must be an integer; invalid: %v", err)
			}
			rl.ZoneSize = size
			fallthrough
		case 2:
			rl.Rate = args[1]
			rl.Key = args[0]
		default:
			return d.ArgErr()
		}

		// Consume the arguments manually since RemainingArgs doesn't
		for i := 0; i < len(args); i++ {
			d.NextArg()
		}

		for d.NextBlock(0) {
			switch d.Val() {
			case "cloudflare":
				rl.Cloudflare = &CloudflareConfig{}
				for nesting := d.Nesting(); d.NextBlock(nesting); {
					switch d.Val() {
					case "api_token":
						if !d.NextArg() {
							return d.ArgErr()
						}
						rl.Cloudflare.APIToken = d.Val()
					case "account_id":
						if !d.NextArg() {
							return d.ArgErr()
						}
						rl.Cloudflare.AccountID = d.Val()
					case "list_id":
						if !d.NextArg() {
							return d.ArgErr()
						}
						rl.Cloudflare.ListID = d.Val()
					default:
						return d.Errf("unknown subdirective: %s", d.Val())
					}
				}
			default:
				return d.Errf("unknown subdirective: %s", d.Val())
			}
		}
	}
	return nil
}

// Interface guards
var (
	_ caddyfile.Unmarshaler = (*RateLimit)(nil)
)
