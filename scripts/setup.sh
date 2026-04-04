#!/bin/bash
set -e

echo "=== FinanceOS Setup ==="

# Check for .env file
if [ ! -f .env ]; then
  if [ -f .env.example ]; then
    cp .env.example .env
    echo "✅ .env criado a partir de .env.example"
    echo "   Edite o arquivo .env com suas credenciais e rode novamente."
    exit 1
  else
    echo "⚠️  .env.example não encontrado. Criando .env básico..."
    cat > .env << 'ENVEOF'
DATABASE_URL=postgresql://financeos:financeos@localhost:5432/financeos
REDIS_URL=redis://localhost:6379
JWT_SECRET=change-me-in-production
APP_ENV=development
APP_PORT=8000
LOG_LEVEL=debug
ANTHROPIC_API_KEY=
EVOLUTION_API_URL=http://localhost:8081
EVOLUTION_API_KEY=
ENVEOF
    echo "✅ .env criado. Edite com suas credenciais e rode novamente."
    exit 1
  fi
fi

echo "✅ .env encontrado"

# Start infrastructure
echo "⏳ Iniciando PostgreSQL e Redis..."
docker-compose up -d postgres redis

# Wait for PostgreSQL
echo "⏳ Aguardando PostgreSQL..."
max_attempts=30
attempt=0
until docker-compose exec -T postgres pg_isready -U financeos 2>/dev/null; do
  attempt=$((attempt + 1))
  if [ $attempt -ge $max_attempts ]; then
    echo "❌ PostgreSQL não respondeu após ${max_attempts} tentativas"
    exit 1
  fi
  printf '.'
  sleep 2
done
echo ""
echo "✅ PostgreSQL pronto"

# Run migrations
echo "⏳ Rodando migrations..."
if command -v migrate &> /dev/null; then
  DB_URL=$(grep DATABASE_URL .env | cut -d '=' -f2-)
  migrate -path packages/database/migrations \
          -database "${DB_URL}?sslmode=disable" \
          up 2>&1 || echo "⚠️  Migrations já aplicadas ou erro (verifique manualmente)"
else
  echo "⚠️  golang-migrate não encontrado. Execute: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
  echo "   Depois rode: make migrate"
fi

echo ""
echo "✅ Setup completo!"
echo ""
echo "▶️  Para iniciar o ambiente de desenvolvimento:"
echo "   docker-compose up -d"
echo ""
echo "▶️  Para iniciar apenas a API localmente:"
echo "   cd apps/api && go run ./cmd/server"
