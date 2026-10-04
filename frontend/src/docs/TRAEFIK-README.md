# Traefik Deployment Guide

## Quick Start

1. **Copy environment file:**
   ```bash
   cp .env.docker.example .env
   ```

2. **Edit .env with your settings:**
   - `DOMAIN` - your domain (e.g., `avtoplaneta.com`)
   - `ACME_EMAIL` - email for Let's Encrypt
   - `DASHBOARD_PASSWORD` - bcrypt hash for dashboard
   - Database credentials
   - API keys

3. **Generate dashboard password:**
   ```bash
   # Linux/Mac
   htpasswd -nbB admin YOUR_PASSWORD
   
   # Or use: https://hostingcanada.org/htpasswd-generator/
   ```

4. **Start services:**
   ```bash
   # Production
   docker compose up -d
   
   # Development (with hot reload)
   docker compose -f docker-compose.yml -f docker-compose.override.yml up
   ```

## Architecture

```
Internet → Traefik (SSL/TLS, ForwardAuth) ──┬──→ Frontend (React, Nginx)
                                            ├──→ Auth Service (8083, ForwardAuth /auth/verify)
                                            ├──→ Parts Service (8081, gRPC-Gateway /api/v1/*, REST)
                                            ├──→ Orders Service (8082, /orders)
                                            ├──→ Messaging Service (8084, /api/messaging/*)
                                            └──→ Export Service (8085, /uploads/pricelist.xml, /api/export/*)
                                                    ↓
                                               Dashboard (traefik.yourdomain.com)
```

## Services

| Service | Port | Description |
|---------|------|-------------|
| Traefik | 80, 443 | Reverse proxy, SSL termination, ForwardAuth middleware |
| Frontend | 80 | React app (nginx) |
| Auth | 8083 / 9083 | Authentication service, ForwardAuth verification (`/auth/verify`), gRPC |
| Parts | 8081 / 9081 | Parts inventory service, gRPC-Gateway (`/api/v1/*`), REST, gRPC |
| Orders | 8082 / 9082 | Orders service, gRPC |
| Messaging | 8084 / 9084 | Messaging service, gRPC |
| Export | 8085 | Export service (pricelist.xml) |
| PostgreSQL | 5432 | Database (isolated per-service or common fallback) |
| Redis | 6379 | Cache / Sessions / Redis Streams events |
| Elasticsearch | 9200 | Search engine |
| Prometheus | 9090 | Metrics collection |
| Grafana | 3000 | Dashboards |

## URLs

- **App**: `https://yourdomain.com`
- **Dashboard**: `https://traefik.yourdomain.com`
- **API (Parts & Inventory)**: `https://yourdomain.com/api/v1/*`, `https://yourdomain.com/api/inventory`
- **Auth**: `https://yourdomain.com/auth/*`
- **Orders**: `https://yourdomain.com/orders/*`
- **Messaging**: `https://yourdomain.com/api/messaging/*`
- **Export**: `https://yourdomain.com/uploads/pricelist.xml`
- **Health**: `https://yourdomain.com/health`
- **Prometheus**: `https://prometheus.yourdomain.com`
- **Grafana**: `https://grafana.yourdomain.com` (admin/admin123)

## Useful Commands

```bash
# View logs
docker compose logs -f traefik
docker compose logs -f auth
docker compose logs -f parts
docker compose logs -f orders
docker compose logs -f prometheus
docker compose logs -f grafana

# Restart service
docker compose restart parts

# Rebuild service
docker compose up -d --build parts

# View Traefik dashboard
open https://traefik.yourdomain.com

# View Prometheus
open https://prometheus.yourdomain.com

# View Grafana
open https://grafana.yourdomain.com

# Check SSL certificates
docker compose exec traefik cat /etc/traefik/acme.json
```

## Firewall (UFW) - Linux

```bash
# Open only HTTP/HTTPS
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# For local access only (internal services):
sudo ufw allow from 127.0.0.1 to any port 5432,6379,9200,9090,3000

# Check status
sudo ufw status
```

## Troubleshooting

### SSL not working
- Check `acme.json` permissions: `chmod 600 traefik/acme.json`
- Verify DNS points to your server
- Check Traefik logs: `docker compose logs traefik`

### Service not reachable
- Check health: `curl https://yourdomain.com/health`
- Check Traefik dashboard for routing
- Verify service is running: `docker compose ps`

### Dashboard auth fails
- Regenerate password hash
- Check `.env` has correct `DASHBOARD_PASSWORD` format
