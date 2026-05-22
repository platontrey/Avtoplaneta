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
Internet → Traefik (SSL/Let's Encrypt) → API Gateway (8080) → Microservices
                                              ↓
                                         Dashboard (traefik.yourdomain.com)
```

## Services

| Service | Port | Description |
|---------|------|-------------|
| Traefik | 80, 443 | Reverse proxy, SSL termination |
| Frontend | 80 | React app (nginx) |
| Gateway | 8080 | API Gateway |
| Auth | 8083 | Authentication service |
| Parts | 8081 | Parts inventory service |
| Orders | 8082 | Orders service |
| Messaging | 8084 | Messaging service |
| PostgreSQL | 5432 | Database |
| Redis | 6379 | Cache/Sessions |
| Elasticsearch | 9200 | Search engine |
| Prometheus | 9090 | Metrics collection |
| Grafana | 3000 | Dashboards |

## URLs

- **App**: `https://yourdomain.com`
- **Dashboard**: `https://traefik.yourdomain.com`
- **API**: `https://yourdomain.com/api/*`
- **Health**: `https://yourdomain.com/health`
- **Prometheus**: `https://prometheus.yourdomain.com`
- **Grafana**: `https://grafana.yourdomain.com` (admin/admin123)

## Useful Commands

```bash
# View logs
docker compose logs -f traefik
docker compose logs -f gateway
docker compose logs -f prometheus
docker compose logs -f grafana

# Restart service
docker compose restart gateway

# Rebuild service
docker compose up -d --build gateway

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
