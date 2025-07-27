# V-Cloud Server Deployment Guide

## Prerequisites

### 1. Server Requirements
- **OS**: Ubuntu 20.04 LTS or later / CentOS 8+ / Debian 10+
- **RAM**: Minimum 2GB (4GB recommended)
- **Storage**: Minimum 10GB free space
- **CPU**: 2 cores minimum
- **Network**: Open ports 80, 443, 8080, 3306 (optional)

### 2. Required Software
- Docker Engine
- Docker Compose
- Git (for cloning repository)
- Optional: Nginx (for reverse proxy)

## Quick Deployment Steps

### Step 1: Server Setup
```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install required packages
sudo apt install -y curl wget git unzip

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Logout and login again to apply Docker group changes
```

### Step 2: Deploy Application
```bash
# Clone or upload your application
git clone <your-repo> vcloud-app
# OR upload files to /opt/vcloud-app

cd vcloud-app

# Make scripts executable
chmod +x setup_docker.sh
chmod +x server_deploy.sh

# Run deployment script
./server_deploy.sh
```

### Step 3: Configure Firewall (Ubuntu/Debian)
```bash
# Enable UFW
sudo ufw enable

# Allow SSH
sudo ufw allow ssh

# Allow HTTP/HTTPS
sudo ufw allow 80
sudo ufw allow 443

# Allow application port
sudo ufw allow 8080

# Optional: Allow direct database access (NOT recommended for production)
# sudo ufw allow 3306

# Check status
sudo ufw status
```

## Environment Configuration

### Production Environment Variables
Create a `.env` file for production settings:

```bash
# Database Configuration
DB_HOST=mysql
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_secure_password_here
DB_NAME=vcloud

# Application Configuration
GIN_MODE=release
APP_ENV=production

# Security
JWT_SECRET=your_jwt_secret_key_here

# File Upload Limits
MAX_UPLOAD_SIZE=100MB
MAX_STORAGE_PER_USER=5GB
```

## SSL/HTTPS Setup with Nginx

### 1. Install Nginx
```bash
sudo apt install nginx certbot python3-certbot-nginx
```

### 2. Configure Nginx (see nginx.conf file)

### 3. Get SSL Certificate
```bash
sudo certbot --nginx -d yourdomain.com
```

## Monitoring and Maintenance

### View Logs
```bash
# Application logs
docker-compose logs -f vcloud-backend

# Database logs
docker-compose logs -f mysql

# All services
docker-compose logs -f
```

### Backup Database
```bash
# Create backup
docker exec vcloud-mysql mysqldump -u root -pvithu vcloud > backup_$(date +%Y%m%d_%H%M%S).sql

# Restore backup
docker exec -i vcloud-mysql mysql -u root -pvithu vcloud < backup_file.sql
```

### Update Application
```bash
# Pull latest changes
git pull origin main

# Rebuild and restart
docker-compose down
docker-compose up --build -d
```

### System Monitoring
```bash
# Check disk usage
df -h

# Check memory usage
free -h

# Check Docker containers
docker ps

# Check application health
curl http://localhost:8080/health
```

## Security Considerations

### 1. Database Security
- Change default passwords
- Disable root remote access
- Use non-root database user for application
- Regular backups

### 2. Application Security
- Use HTTPS in production
- Configure CORS properly
- Implement rate limiting
- Regular security updates

### 3. Server Security
- Keep system updated
- Configure firewall
- Use SSH keys instead of passwords
- Regular security audits

## Troubleshooting

### Common Issues

#### Database Connection Failed
```bash
# Check if MySQL container is running
docker ps | grep mysql

# Check MySQL logs
docker-compose logs mysql

# Test database connection
docker exec -it vcloud-mysql mysql -u root -pvithu -e "SHOW DATABASES;"
```

#### Application Won't Start
```bash
# Check application logs
docker-compose logs vcloud-backend

# Check if port is available
sudo netstat -tlnp | grep :8080

# Restart services
docker-compose restart
```

#### Storage Issues
```bash
# Check volume mounts
docker volume ls
docker volume inspect vcloud_uploads_data

# Check disk space
df -h

# Clean up Docker
docker system prune -a
```

## Production Checklist

- [ ] Change all default passwords
- [ ] Configure SSL/HTTPS
- [ ] Set up automated backups
- [ ] Configure monitoring
- [ ] Test disaster recovery
- [ ] Document access credentials
- [ ] Set up log rotation
- [ ] Configure fail2ban
- [ ] Test all functionality
- [ ] Set up monitoring alerts

## Performance Optimization

### Database Optimization
```sql
-- Add these to MySQL configuration
innodb_buffer_pool_size = 1G
innodb_log_file_size = 256M
max_connections = 200
```

### Application Optimization
- Use connection pooling
- Implement caching
- Optimize file upload handling
- Monitor resource usage

## Backup Strategy

### Automated Backup Script
```bash
#!/bin/bash
# Create in /opt/vcloud-backup.sh

BACKUP_DIR="/opt/backups"
DATE=$(date +%Y%m%d_%H%M%S)

# Database backup
docker exec vcloud-mysql mysqldump -u root -pvithu vcloud > $BACKUP_DIR/db_$DATE.sql

# Compress files
tar -czf $BACKUP_DIR/userfiles_$DATE.tar.gz ./volumes/userfiles/

# Keep only last 7 days
find $BACKUP_DIR -name "*.sql" -mtime +7 -delete
find $BACKUP_DIR -name "*.tar.gz" -mtime +7 -delete
```

Add to crontab:
```bash
# Daily backup at 2 AM
0 2 * * * /opt/vcloud-backup.sh
```
