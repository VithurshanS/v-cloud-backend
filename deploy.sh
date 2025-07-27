#!/bin/bash

# V-Cloud Complete Setup Script
# This script installs Docker, sets up the environment, and deploys the V-Cloud application

set -e  # Exit on any error

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

echo "========================================"
echo "V-Cloud Complete Setup & Deployment"
echo "========================================"

# Check if running as root
if [[ $EUID -eq 0 ]]; then
   print_error "This script should not be run as root for security reasons"
   exit 1
fi

# Update system packages
print_step "Updating system packages..."
sudo apt update && sudo apt upgrade -y

# Install required packages
print_step "Installing required packages..."
sudo apt install -y curl wget git unzip software-properties-common

# Install Docker if not exists
print_step "Checking Docker installation..."
if ! command -v docker &> /dev/null; then
    print_status "Installing Docker..."
    curl -fsSL https://get.docker.com -o get-docker.sh
    sudo sh get-docker.sh
    sudo usermod -aG docker $USER
    rm get-docker.sh
    print_status "Docker installed successfully"
else
    print_status "Docker is already installed"
fi

# Install Docker Compose if not exists
print_step "Checking Docker Compose installation..."
if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    print_status "Installing Docker Compose..."
    sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    sudo chmod +x /usr/local/bin/docker-compose
    print_status "Docker Compose installed successfully"
else
    print_status "Docker Compose is already available"
fi

# Check if user is in docker group and can access Docker
print_step "Checking Docker permissions..."
if groups $USER | grep &>/dev/null '\bdocker\b'; then
    # User is in docker group, but check if we can actually use Docker
    if docker info >/dev/null 2>&1; then
        USE_SUDO=""
        print_status "Docker permissions OK"
    else
        print_warning "User is in docker group but needs to logout/login for changes to take effect"
        print_warning "Using sudo for now..."
        USE_SUDO="sudo"
    fi
else
    print_warning "User $USER is not in docker group. Adding to group..."
    sudo usermod -aG docker $USER
    print_warning "You will need to logout and login again for Docker group changes to take effect."
    print_warning "Using sudo for now..."
    USE_SUDO="sudo"
fi

# Create volume directories with proper permissions
print_step "Creating volume directories..."
mkdir -p ./volumes/uploads ./volumes/userfiles ./volumes/mysql

# Set proper ownership and permissions - make volumes accessible to container user (1001)
# This ensures the container can read/write to mounted volumes
print_status "Setting directory ownership and permissions..."
sudo chown -R 1001:1001 ./volumes/uploads ./volumes/userfiles
sudo chmod -R 755 ./volumes/uploads ./volumes/userfiles
print_status "Volume directories created and permissions set."

# Configure firewall (optional)
read -p "Do you want to configure firewall? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    print_step "Configuring firewall..."
    sudo ufw --force enable
    sudo ufw allow ssh
    sudo ufw allow 80
    sudo ufw allow 443
    sudo ufw allow 8080
    sudo ufw allow 3307  # MySQL external port
    print_status "Firewall configured"
fi

# Stop any existing containers
print_step "Stopping any existing containers..."
$USE_SUDO docker-compose down 2>/dev/null || true

# Pull latest images and build
print_step "Building and starting V-Cloud services..."
$USE_SUDO docker-compose up --build -d

# Wait for services to be ready
print_step "Waiting for services to start..."
sleep 30

# Health check
print_step "Performing health checks..."
max_attempts=30
attempt=1

while [ $attempt -le $max_attempts ]; do
    if curl -f http://localhost:8080/health >/dev/null 2>&1; then
        print_status "Application health check passed ✓"
        break
    else
        if [ $attempt -eq $max_attempts ]; then
            print_error "Health check failed after $max_attempts attempts"
            print_error "Check logs with: docker-compose logs"
            echo ""
            print_step "Showing recent logs..."
            $USE_SUDO docker-compose logs --tail=20
            exit 1
        fi
        print_warning "Health check attempt $attempt/$max_attempts failed, retrying in 5 seconds..."
        sleep 5
        ((attempt++))
    fi
done

# Create backup script
print_step "Creating maintenance scripts..."
cat > vcloud-backup.sh << 'EOF'
#!/bin/bash
BACKUP_DIR="./backups"
DATE=$(date +%Y%m%d_%H%M%S)

echo "Creating backup at $DATE"

# Database backup
docker exec vcloud-mysql mysqladmin -u root -pvithu processlist > /dev/null 2>&1
if [ $? -eq 0 ]; then
    docker exec vcloud-mysql mysqldump -u root -pvithu vcloud > $BACKUP_DIR/db_$DATE.sql
    echo "Database backup created: $BACKUP_DIR/db_$DATE.sql"
else
    echo "Error: Cannot connect to database"
    exit 1
fi

# Compress user files
if [ -d "./volumes/userfiles" ]; then
    tar -czf $BACKUP_DIR/userfiles_$DATE.tar.gz ./volumes/userfiles/ 2>/dev/null || true
    echo "User files backup created: $BACKUP_DIR/userfiles_$DATE.tar.gz"
fi

# Keep only last 7 days
find $BACKUP_DIR -name "*.sql" -mtime +7 -delete 2>/dev/null || true
find $BACKUP_DIR -name "*.tar.gz" -mtime +7 -delete 2>/dev/null || true

echo "Backup completed successfully!"
EOF

chmod +x vcloud-backup.sh

# Create maintenance script
cat > vcloud-manage.sh << 'EOF'
#!/bin/bash

COMPOSE_CMD="docker-compose"
if groups $USER | grep &>/dev/null '\bdocker\b'; then
    USE_SUDO=""
else
    USE_SUDO="sudo"
fi

case "$1" in
    "logs")
        $USE_SUDO $COMPOSE_CMD logs -f
        ;;
    "status")
        $USE_SUDO $COMPOSE_CMD ps
        ;;
    "restart")
        $USE_SUDO $COMPOSE_CMD restart
        ;;
    "stop")
        $USE_SUDO $COMPOSE_CMD down
        ;;
    "start")
        $USE_SUDO $COMPOSE_CMD up -d
        ;;
    "update")
        echo "Updating application..."
        $USE_SUDO $COMPOSE_CMD down
        $USE_SUDO $COMPOSE_CMD up --build -d
        ;;
    "backup")
        ./vcloud-backup.sh
        ;;
    "cleanup")
        $USE_SUDO docker system prune -f
        ;;
    "health")
        curl -s http://localhost:8080/health | jq . 2>/dev/null || curl -s http://localhost:8080/health
        ;;
    *)
        echo "V-Cloud Management Script"
        echo "Usage: $0 {logs|status|restart|stop|start|update|backup|cleanup|health}"
        echo ""
        echo "Commands:"
        echo "  logs     - View application logs"
        echo "  status   - Show container status"
        echo "  restart  - Restart all services"
        echo "  stop     - Stop all services"
        echo "  start    - Start all services"
        echo "  update   - Update and restart application"
        echo "  backup   - Create database and files backup"
        echo "  cleanup  - Clean up Docker resources"
        echo "  health   - Check application health"
        ;;
esac
EOF

chmod +x vcloud-manage.sh

# Display final status
print_step "Checking final status..."
$USE_SUDO docker-compose ps

echo ""
echo "========================================"
echo -e "${GREEN}🎉 V-Cloud Deployment Complete!${NC}"
echo "========================================"
echo ""
echo "🌐 Application URLs:"
echo "   V-Cloud Backend:  http://$(hostname -I | awk '{print $1}'):8080"
echo "   Health Check:     http://$(hostname -I | awk '{print $1}'):8080/health"
echo "   phpMyAdmin:       http://$(hostname -I | awk '{print $1}'):8081"
echo ""
echo "📁 Important Files:"
echo "   Backup Script:    ./vcloud-backup.sh"
echo "   Management:       ./vcloud-manage.sh"
echo "   Volumes:          ./volumes/"
echo "   Backups:          ./backups/"
echo ""
echo "🔧 Quick Commands:"
echo "   View logs:        ./vcloud-manage.sh logs"
echo "   Check status:     ./vcloud-manage.sh status"
echo "   Restart:          ./vcloud-manage.sh restart"
echo "   Create backup:    ./vcloud-manage.sh backup"
echo "   Check health:     ./vcloud-manage.sh health"
echo ""
echo "📊 Service Status:"
echo "   MySQL Database:   Port 3307 (external)"
echo "   V-Cloud API:      Port 8080"
echo "   phpMyAdmin:       Port 8081"
echo ""

# Test the application
echo "🧪 Testing application..."
if curl -s http://localhost:8080/health | grep -q "healthy"; then
    echo -e "${GREEN}✅ Application is running and healthy!${NC}"
else
    echo -e "${YELLOW}⚠️  Application may not be fully ready yet. Check logs if needed.${NC}"
fi

echo ""
echo -e "${GREEN}🚀 Setup completed successfully!${NC}"
echo ""
echo "Next steps:"
echo "1. Test your API endpoints"
echo "2. Set up regular backups with cron"
echo "3. Configure SSL/HTTPS for production"
echo ""

if ! groups $USER | grep &>/dev/null '\bdocker\b'; then
    echo -e "${YELLOW}⚠️  IMPORTANT: You need to logout and login again for Docker group changes to take effect.${NC}"
    echo "   After relogging, you won't need sudo for Docker commands."
fi
