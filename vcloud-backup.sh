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
