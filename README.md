# Caddy Cloudflare Rate Limit (CFRCL) Plugin

Caddy Cloudflare Rate Limit (CFRCL) is a Caddy module that provides advanced rate-limiting capabilities by integrating with Cloudflare IP Lists.

When a client exceeds a configured rate limit, their IP address is automatically added to a specified Cloudflare IP List for a defined duration. This allows you to leverage Cloudflare's WAF and firewall rules to enforce rate limits at the edge, protecting your upstream services from abusive traffic.

## Features

- **IP-based Rate Limiting**: Configure requests per second and burst limits per client IP.
- **Cloudflare IP List Integration**: Automatically add/remove offending IPs to/from a Cloudflare IP List.
- **Configurable Block Duration**: Specify how long an IP remains blocked in the Cloudflare IP List.
- **Caddyfile Support**: Easy configuration directly in your Caddyfile.

## Installation

To use this plugin, you need to build Caddy with `xcaddy`. Make sure you have `xcaddy` installed (if not, `go install github.com/caddyserver/xcaddy/cmd/xcaddy@latest`).

```bash
xcaddy build --with github.com/HellDarkK/caddy-cfrl
```

This command will produce a Caddy binary with the `cloudflare_ratelimit` module included.

## Configuration

To use the `cloudflare_ratelimit` directive, add it to your Caddyfile. The directive should generally be placed *before* any `reverse_proxy` or other directives that handle upstream requests, to ensure rate limiting occurs first.

Here's an example Caddyfile configuration:

```caddyfile
{
    # It's recommended to order cloudflare_ratelimit before reverse_proxy
    order cloudflare_ratelimit before reverse_proxy
}

:8080 {
    route {
        cloudflare_ratelimit {
            api_token "YOUR_CLOUDFLARE_API_TOKEN"  # Your Cloudflare API Token
            account_id "YOUR_ACCOUNT_ID"          # Your Cloudflare Account ID
            list_id "YOUR_LIST_ID"               # The ID of the Cloudflare IP List to use
            rate 1.0                               # Average requests per second (float)
            burst 5                                # Maximum burst size (int)
            entry_duration 5m                      # How long to block the IP (e.g., 1m, 30s, 1h)
        }
        
        # Your backend service
        reverse_proxy localhost:9000

        # Example: Respond directly if no reverse_proxy is needed
        # respond "Hello, world! You are not rate-limited."
    }
}
```

### Directive Syntax

```
cloudflare_ratelimit {
    api_token <token>
    account_id <id>
    list_id <id>
    rate <float>
    burst <int>
    entry_duration <duration>
}
```

### Options

-   `api_token` (required): Your Cloudflare API Token. This token needs permissions to read and write to IP Lists.
-   `account_id` (required): Your Cloudflare Account ID. You can find this in your Cloudflare dashboard.
-   `list_id` (required): The ID of the Cloudflare IP List where blocked IPs will be added. You must create this list manually in your Cloudflare dashboard beforehand.
-   `rate` (required): The average number of requests per second allowed before an IP is blocked (e.g., `1.0` for 1 request per second).
-   `burst` (required): The maximum burst of requests allowed before rate limiting takes effect. This allows for occasional spikes in traffic without immediate blocking.
-   `entry_duration` (required): The duration for which an IP will remain in the Cloudflare IP List after being blocked (e.g., `1m` for 1 minute, `30s` for 30 seconds, `1h` for 1 hour).

## How it Works

1.  **Rate Limiting**: For each incoming request, the plugin identifies the client's IP address and checks it against an in-memory rate limiter.
2.  **Block IP**: If the rate limit for an IP is exceeded:
    *   The request is immediately responded to with a `429 Too Many Requests` status.
    *   The client's IP address is asynchronously added to the configured Cloudflare IP List using the Cloudflare API.
3.  **Unblock IP**: After the `entry_duration` expires, the plugin asynchronously removes the IP address from the Cloudflare IP List.

This setup offloads the actual IP blocking to Cloudflare's network edge, reducing the load on your Caddy server and providing distributed protection.

## Development

To contribute or develop this module:

1.  Clone the repository:
    ```bash
    git clone https://github.com/HellDarkK/caddy-cfrl.git
    cd caddy-cfrl
    ```
2.  Run tests (if any are implemented):
    ```bash
    go test ./...
    ```
3.  Build with `xcaddy` as described in the Installation section.

## License

This project is licensed under the MIT License - see the `LICENSE` file for details. (Note: A `LICENSE` file should be added to the project if not already present.)
