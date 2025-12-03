#!/usr/bin/env bash
set -euo pipefail

# ============================================================================
# Database Management Script
# ============================================================================
# Manages local databases using Docker Compose
# Usage: ./scripts/db.sh [command]
# ============================================================================

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m'

print_usage() {
  cat << EOF
${BOLD}Database Management${NC}

${BLUE}Usage:${NC}
  ./scripts/db.sh [command]

${BLUE}Commands:${NC}
  ${BOLD}up${NC}              Start all databases (PostgreSQL + MongoDB)
  ${BOLD}down${NC}            Stop all databases
  ${BOLD}restart${NC}         Restart all databases
  ${BOLD}logs${NC}            Show database logs
  ${BOLD}ps${NC}              Show running containers
  ${BOLD}postgres:shell${NC}  Connect to PostgreSQL CLI
  ${BOLD}mongo:shell${NC}     Connect to MongoDB shell
  ${BOLD}postgres:dump${NC}   Dump PostgreSQL database
  ${BOLD}postgres:restore${NC} Restore PostgreSQL database
  ${BOLD}clean${NC}           Remove all containers and volumes

${BLUE}Examples:${NC}
  ./scripts/db.sh up
  ./scripts/db.sh postgres:shell
  ./scripts/db.sh down
EOF
}

# Check if docker-compose is available
if ! command -v docker-compose &> /dev/null; then
  echo "${RED}❌ docker-compose is not installed${NC}"
  echo "   Install Docker Desktop or docker-compose to use this script"
  exit 1
fi

# Create docker-compose.yml if it doesn't exist
create_docker_compose() {
  if [ -f "docker-compose.yml" ]; then
    return
  fi
  
  echo "${BLUE}📝 Creating docker-compose.yml...${NC}"
  
  cat > docker-compose.yml << 'COMPOSE_EOF'
version: '3.8'

services:
  postgres:
    image: postgres:16-alpine
    container_name: go-monolith-postgres
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: appdb
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5

  mongodb:
    image: mongo:7.0
    container_name: go-monolith-mongodb
    environment:
      MONGO_INITDB_ROOT_USERNAME: admin
      MONGO_INITDB_ROOT_PASSWORD: admin
      MONGO_INITDB_DATABASE: appdb
    ports:
      - "27017:27017"
    volumes:
      - mongodb_data:/data/db
    healthcheck:
      test: ["CMD", "mongosh", "--eval", "db.adminCommand('ping')"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  postgres_data:
  mongodb_data:
COMPOSE_EOF
  
  echo "   ✅ docker-compose.yml created"
fi

# Main command handling
case "${1:-help}" in
  up)
    echo "${BLUE}📦 Starting databases...${NC}"
    create_docker_compose
    docker-compose up -d
    echo ""
    echo "${GREEN}✅ Databases started${NC}"
    echo ""
    echo "${BLUE}Connection strings:${NC}"
    echo "  PostgreSQL: postgresql://postgres:postgres@localhost:5432/appdb"
    echo "  MongoDB:    mongodb://admin:admin@localhost:27017/appdb?authSource=admin"
    echo ""
    echo "Waiting for databases to be ready..."
    sleep 5
    docker-compose ps
    ;;
  
  down)
    echo "${BLUE}🛑 Stopping databases...${NC}"
    if [ -f "docker-compose.yml" ]; then
      docker-compose down
      echo "${GREEN}✅ Databases stopped${NC}"
    else
      echo "${YELLOW}⚠️  docker-compose.yml not found${NC}"
    fi
    ;;
  
  restart)
    echo "${BLUE}🔄 Restarting databases...${NC}"
    create_docker_compose
    docker-compose restart
    echo "${GREEN}✅ Databases restarted${NC}"
    sleep 5
    docker-compose ps
    ;;
  
  logs)
    echo "${BLUE}📋 Database logs:${NC}"
    if [ -f "docker-compose.yml" ]; then
      docker-compose logs -f
    else
      echo "${RED}❌ docker-compose.yml not found${NC}"
      exit 1
    fi
    ;;
  
  ps)
    echo "${BLUE}📊 Running containers:${NC}"
    if [ -f "docker-compose.yml" ]; then
      docker-compose ps
    else
      echo "${RED}❌ docker-compose.yml not found${NC}"
      exit 1
    fi
    ;;
  
  postgres:shell)
    echo "${BLUE}🗄️  Connecting to PostgreSQL...${NC}"
    if command -v psql &> /dev/null; then
      psql -h localhost -U postgres -d appdb
    else
      echo "${YELLOW}⚠️  psql not found, using docker exec...${NC}"
      docker exec -it go-monolith-postgres psql -U postgres -d appdb
    fi
    ;;
  
  mongo:shell)
    echo "${BLUE}🗄️  Connecting to MongoDB...${NC}"
    if command -v mongosh &> /dev/null; then
      mongosh "mongodb://admin:admin@localhost:27017/appdb?authSource=admin"
    else
      echo "${YELLOW}⚠️  mongosh not found, using docker exec...${NC}"
      docker exec -it go-monolith-mongodb mongosh -u admin -p admin --authenticationDatabase admin appdb
    fi
    ;;
  
  postgres:dump)
    BACKUP_FILE="backups/postgres_$(date +%Y%m%d_%H%M%S).sql"
    mkdir -p backups
    echo "${BLUE}💾 Dumping PostgreSQL database to $BACKUP_FILE...${NC}"
    docker exec go-monolith-postgres pg_dump -U postgres appdb > "$BACKUP_FILE"
    echo "${GREEN}✅ Database dumped: $BACKUP_FILE${NC}"
    ;;
  
  postgres:restore)
    if [ -z "${2:-}" ]; then
      echo "${RED}❌ Usage: ./scripts/db.sh postgres:restore <backup_file>${NC}"
      exit 1
    fi
    
    BACKUP_FILE="$2"
    if [ ! -f "$BACKUP_FILE" ]; then
      echo "${RED}❌ Backup file not found: $BACKUP_FILE${NC}"
      exit 1
    fi
    
    echo "${BLUE}📥 Restoring PostgreSQL database from $BACKUP_FILE...${NC}"
    docker exec -i go-monolith-postgres psql -U postgres appdb < "$BACKUP_FILE"
    echo "${GREEN}✅ Database restored${NC}"
    ;;
  
  clean)
    echo "${RED}🗑️  Removing all containers and volumes...${NC}"
    echo "${YELLOW}⚠️  This will delete all data!${NC}"
    read -p "Are you sure? (type 'yes' to confirm): " -r confirm
    if [ "$confirm" = "yes" ]; then
      if [ -f "docker-compose.yml" ]; then
        docker-compose down -v
        echo "${GREEN}✅ Cleanup complete${NC}"
      else
        echo "${RED}❌ docker-compose.yml not found${NC}"
        exit 1
      fi
    else
      echo "Aborted."
    fi
    ;;
  
  help|--help|-h)
    print_usage
    ;;
  
  *)
    echo "${RED}❌ Unknown command: $1${NC}"
    echo ""
    print_usage
    exit 1
    ;;
esac
