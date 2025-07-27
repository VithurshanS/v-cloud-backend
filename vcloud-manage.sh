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
