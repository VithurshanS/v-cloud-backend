# V-Cloud Quick Start Guide

## Simple One-Command Setup

```bash
# For both local development and server deployment
chmod +x deploy.sh && ./deploy.sh
```

This single script will:
- Install Docker and Docker Compose (if not already installed)
- Create necessary directories and volumes
- Build and start all services
- Perform health checks
- Create management scripts

## Manual Setup (Alternative)

### For Local Development
```bash
# 1. Clone/Download the project
# 2. Navigate to project directory
cd V-Cloud-backend

# 3. Run deployment script
./deploy.sh
```

### For Server Deployment
```bash
# 1. Upload all files to your server
scp -r . user@your-server:/opt/vcloud/

# 2. SSH to your server
ssh user@your-server

# 3. Run deployment
cd /opt/vcloud && ./deploy.sh
```

## Available Scripts

| Script | Purpose |
|--------|---------|
| `deploy.sh` | Complete setup and deployment script |
| `vcloud-manage.sh` | Application management commands (created by deploy.sh) |
| `vcloud-backup.sh` | Backup database and files (created by deploy.sh) |

## Ports and Services

| Service | Local Port | Description |
|---------|------------|-------------|
| V-Cloud Backend | 8080 | Main application API |
| MySQL Database | 3307 | Database (external access) |
| phpMyAdmin | 8081 | Database management UI |

## Environment Files

- `.env` - Production environment variables
- `docker-compose.yml` - Main Docker configuration
- `docker-compose.prod.yml` - Production overrides
- `nginx.conf` - Nginx reverse proxy configuration

## Quick Commands

```bash
# Management commands (available after running deploy.sh)
./vcloud-manage.sh logs      # View logs
./vcloud-manage.sh status    # Check status
./vcloud-manage.sh restart   # Restart services
./vcloud-manage.sh backup    # Create backup
./vcloud-manage.sh health    # Check application health
./vcloud-manage.sh update    # Update application
```

## Security Notes

1. Change default passwords in `.env`
2. Configure firewall properly
3. Set up SSL with Let's Encrypt
4. Regular backups and updates
5. Monitor logs for suspicious activity

## Troubleshooting

### Application won't start
```bash
docker-compose logs vcloud-backend
```

### Database connection issues
```bash
docker-compose logs mysql
```

### Check if services are running
```bash
docker-compose ps
```

### Test application health
```bash
curl http://localhost:8080/health
```
