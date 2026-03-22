# Declarative Reverse Proxy Plan

## Problem
Nginx Proxy Manager requires manual GUI interaction to add proxy hosts and request SSL certificates. We need a declarative, config-as-code approach.

## Option Comparison

| Feature | Caddy | Traefik |
|---|---|---|
| Auto HTTPS (Let's Encrypt) | Built-in, zero config | Built-in, needs config |
| Config format | Caddyfile (simple) or JSON | YAML labels or file provider |
| Docker integration | Via docker-proxy or file config | Native via container labels |
| Complexity | Very low | Medium |
| Reload on config change | Yes (API or file watch) | Yes (file watch or Docker events) |
| Dashboard/GUI | None (not needed) | Optional built-in dashboard |

## Recommendation: Caddy

Caddy is the best fit because:
- **Automatic HTTPS by default** — just specify a domain and it handles Let's Encrypt/ZeroSSL
- **Simplest config** — a Caddyfile is 2-3 lines per service
- **No GUI needed** — everything is declarative
- **Hot reload** — config changes apply without downtime
- **Handles cert renewal** automatically

## Architecture

```
                  ┌─────────────────────────┐
    :80/:443 ───► │  Caddy (reverse proxy)  │
                  └──────────┬──────────────┘
                             │
              ┌──────────────┼──────────────┐
              ▼              ▼              ▼
          openbao:8200   portainer:9000   app:PORT
```

All services communicate over a shared Docker network. Caddy is the only container exposing ports 80/443 to the host.

## Implementation

### 1. Directory Structure

```
deploy/proxy/
├── docker-compose.yml
├── Caddyfile            # proxy rules (declarative config)
└── data/                # auto-created: certs, OCSP staples
```

### 2. docker-compose.yml

```yaml
services:
  caddy:
    image: caddy:2-alpine
    container_name: caddy
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
      - "443:443/udp"  # HTTP/3
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile:ro
      - caddy_data:/data       # certs and keys
      - caddy_config:/config   # runtime config
    networks:
      - proxy

networks:
  proxy:
    name: proxy
    external: true

volumes:
  caddy_data:
  caddy_config:
```

### 3. Caddyfile (example)

```caddyfile
# Global options
{
    email it.mvha@gmail.com
}

# OpenBao
obv.vhco.pro {
    reverse_proxy openbao:8200
}

# Portainer
portainer.vhco.pro {
    reverse_proxy portainer:9000
}

# Add more services — just append a block:
# app.vhco.pro {
#     reverse_proxy myapp:3000
# }
```

That's it. Each service is 3 lines. Caddy auto-requests and renews certs for every domain listed.

### 4. Network Setup

Other compose stacks need to join the shared `proxy` network so Caddy can reach them:

```yaml
# In deploy/openbao/docker-compose.yml, add:
services:
  openbao:
    networks:
      - default
      - proxy

networks:
  proxy:
    external: true
```

### 5. Migration Steps

1. Create the shared Docker network: `docker network create proxy`
2. Deploy Caddy with the Caddyfile
3. Update existing compose files to join the `proxy` network
4. Verify DNS records point to this server for each domain
5. `docker compose up -d` — Caddy auto-requests certs on first start
6. Confirm HTTPS works, then tear down NPM: `cd ~/npm && docker compose down`

### 6. Day-2 Operations

- **Add a service**: Add a new block to `Caddyfile`, run `docker exec caddy caddy reload --config /etc/caddy/Caddyfile`
- **Remove a service**: Delete the block, reload
- **All config is in git** — no GUI state to back up or lose

## Prerequisites

- DNS A records for each domain must point to this server's public IP (130.61.216.36)
- Ports 80 and 443 must be open inbound (already are, since NPM uses them)

## Notes

- Caddy stores cert state in a Docker volume (`caddy_data`), so certs survive container restarts
- If you ever need DNS-01 challenges (wildcard certs), Caddy supports them via plugins (e.g., `caddy-dns/cloudflare`)
- The Caddyfile can be templated or generated from a YAML file if you want an even more structured approach later
