# ratelimit

A Caddy v2 extension to apply IP-based rate-limiting for HTTP requests. Optionally, block offending IPs via a Cloudflare Custom List when the rate limit is exceeded.

## Features

- Powerful and configurable rate limiting by IP or request property (see placeholders).
- **Cloudflare Integration:** Automatically add violating client IPs to a specified Cloudflare IP Custom List via the Cloudflare API (optional).

---

## Installation

```
$ xcaddy build --with github.com/HellDarkK/caddy-cfrl/ratelimit
```

---

## Caddyfile Syntax

```
rate_limit [<matcher>] <key> <rate> [<zone_size> [<reject_status>]] {
    cloudflare {
        api_token <api_token>
        account_id <account_id>
        list_id   <list_id>
    }
}
```

**Parameters:**
- `<key>`: The variable to uniquely identify a client. Supported variables ([Caddy shorthand placeholders][1]):
    + `{path.<var>}`
    + `{query.<var>}`
    + `{header.<VAR>}`
    + `{cookie.<var>}`
    + `{body.<var>}` (requires the [requestbodyvar](https://github.com/HellDarkK/caddy-cfrl/tree/master/requestbodyvar) extension)
    + `{remote.host}` (ignores the `X-Forwarded-For` header)
    + `{remote.port}`
    + `{remote.ip}` (prefers the first IP in the `X-Forwarded-For` header)
    + `{remote.host_prefix.<bits>}` (CIDR block version)
    + `{remote.ip_prefix.<bits>}` (CIDR block version)
- `<rate>`: The request rate limit (per client key) as `Nr/s` or `Nr/m` (requests per second/minute).
- `<zone_size>`: Max number of key values (LRU zone).
- `<reject_status>`: HTTP status code to return when limited (default: 429).

---

## Example: Basic Rate Limiting

```
localhost:8080 {
    route /foo {
        rate_limit {remote.ip} 2r/m
        respond 200
    }
}
```

Limits `/foo` route to 2 requests/minute per IP. Exceeding the limit gets a 429 response.

---

## Example: With Cloudflare Integration

```
localhost:8080 {
    route /api {
        rate_limit {remote.ip} 5r/m {
            cloudflare {
                api_token <YOUR_API_TOKEN>
                account_id <YOUR_ACCOUNT_ID>
                list_id   <YOUR_LIST_ID>
            }
        }
        respond 200
    }
}
```
- When a client exceeds the specified rate, their IP will be added to the specified Cloudflare List.
- The Cloudflare API call is asynchronous and errors are logged to the Caddy log.

---

## Cloudflare Setup Guide

1. **Create an IP Custom List:**
   - Go to [Cloudflare Dashboard](https://dash.cloudflare.com/).
   - Choose your account, go to **Lists** > **IP Lists** > **Create List**.
   - Give the list a name and note the `List ID` after creation.
2. **Generate an API Token:**
   - Go to **My Profile** > **API Tokens** > **Create Token**.
   - Required permissions:
     - `Zone > Lists: Edit`
     - `Account > Lists: Edit`
   - Scope the token as narrowly as possible.
   - Keep your API token secret.
3. **Find Your Account ID:**
   - Account ID is in the dashboard URL or account settings.

---

## Security & Operational Notes

- Cloudflare integration is optional; omit the `cloudflare` block if you don't want to block list IPs.
- If Cloudflare API calls fail, requests are still rate-limited; blocking is best effort and errors are logged.
- Rate limiting can be evaded by IP rotation if `{remote.ip}` is used. Use other keys as needed.
- Ensure you comply with Cloudflare List API rate limits [see docs](https://developers.cloudflare.com/api/operations/account-lists-create-list-item/).

---

[1]: https://caddyserver.com/docs/caddyfile/concepts#placeholders

