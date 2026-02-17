# caddy-cfrl

**Focused Extensions for Caddy v2**

This repository now contains two modules:

- **ratelimit:**
    - Caddy HTTP middleware for configurable rate limiting by IP (or other properties).
    - Optionally, when a rate limit is exceeded, the offending client IP is added to a Cloudflare "Custom List" via the Cloudflare API. That IP will be automatically removed from the list 20 seconds later. This allows integrated, fast, self-healing blocking.
- **cloudflare:**
    - Lightweight helper package for interacting with the Cloudflare List API.
    - Used internally by ratelimit for blocking IPs.

All other extensions have been removed. The codebase is highly focused on application-level rate limiting and Cloudflare threat mitigation.

---

## Installation

Build using [xcaddy](https://github.com/caddyserver/xcaddy):

```
$ xcaddy build --with github.com/HellDarkK/caddy-cfrl
```

---

## Usage

See [ratelimit/README.md](ratelimit/README.md) for complete documentation, configuration, and examples including Cloudflare integration.

---

## License

[MIT](LICENSE)

